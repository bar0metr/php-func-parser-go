package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"php-func-parser-go/internal/config"
	"php-func-parser-go/internal/fs"
	"php-func-parser-go/internal/model"
	"php-func-parser-go/internal/parser"
	"php-func-parser-go/internal/report"
)

type App struct {
	cfg    config.Config
	logger *slog.Logger

	finder fs.Finder
	parser parser.FunctionParser
	writer report.Writer
}

func New(cfg config.Config, logger *slog.Logger) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
		finder: fs.NewOSFinder(),
		parser: parser.NewPHPFunctionParser(),
		writer: report.NewTextWriter(),
	}
}

func (a *App) Run(ctx context.Context) error {
	if a.cfg.DiffOld != "" || a.cfg.DiffNew != "" {
		return a.runDiff(ctx)
	}

	if a.logger.Enabled(ctx, slog.LevelDebug) {
		a.logger.Debug(
			"starting",
			"path", a.cfg.Path,
			"recursive", a.cfg.Recursive,
			"report", outputName(a.cfg.Report),
			"workers", a.cfg.Workers,
		)
	}

	files, err := a.finder.PHPFiles(ctx, a.cfg.Path, a.cfg.Recursive)
	if err != nil {
		return fmt.Errorf("listing PHP files: %w", err)
	}
	if a.logger.Enabled(ctx, slog.LevelDebug) {
		a.logger.Debug("php files discovered", "count", len(files))
	}

	if len(files) == 0 {
		a.logger.Info("no .php files found", "path", a.cfg.Path)
		return nil
	}

	workers := a.cfg.Workers
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0) * 2
		if workers < 2 {
			workers = 2
		}
	}
	if a.logger.Enabled(ctx, slog.LevelDebug) {
		a.logger.Debug("worker pool configured", "workers", workers)
	}

	reports, err := a.parseFiles(ctx, files, workers)
	if err != nil {
		return err
	}
	sort.Slice(reports, func(i, j int) bool { return reports[i].Path < reports[j].Path })

	out, closeFn, err := a.openOutput(ctx, a.cfg.Report)
	if err != nil {
		return err
	}
	defer closeFn()

	if err := a.writer.Write(out, reports); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}
	a.logger.Info("report written", "files", len(reports), "output", outputName(a.cfg.Report))
	return nil
}

func (a *App) runDiff(ctx context.Context) error {
	if a.logger.Enabled(ctx, slog.LevelDebug) {
		a.logger.Debug(
			"starting diff",
			"diff_old", a.cfg.DiffOld,
			"diff_new", a.cfg.DiffNew,
			"report", outputName(a.cfg.Report),
		)
	}

	oldF, err := os.Open(a.cfg.DiffOld)
	if err != nil {
		return fmt.Errorf("open diff-old: %w", err)
	}
	defer func() { _ = oldF.Close() }()
	newF, err := os.Open(a.cfg.DiffNew)
	if err != nil {
		return fmt.Errorf("open diff-new: %w", err)
	}
	defer func() { _ = newF.Close() }()

	oldReports, err := report.ReadText(oldF)
	if err != nil {
		return fmt.Errorf("read diff-old report: %w", err)
	}
	newReports, err := report.ReadText(newF)
	if err != nil {
		return fmt.Errorf("read diff-new report: %w", err)
	}

	oldIndex := buildSignatureIndex(oldReports)
	diffReports := diffOnlyNew(newReports, oldIndex)

	out, closeFn, err := a.openOutput(ctx, a.cfg.Report)
	if err != nil {
		return err
	}
	defer closeFn()

	if err := a.writer.Write(out, diffReports); err != nil {
		return fmt.Errorf("writing diff report: %w", err)
	}
	a.logger.Info("diff report written", "files", len(diffReports), "output", outputName(a.cfg.Report))
	return nil
}

func buildSignatureIndex(reports []model.FileReport) map[string]map[string]struct{} {
	idx := make(map[string]map[string]struct{}, len(reports))
	for _, r := range reports {
		m := idx[r.Path]
		if m == nil {
			m = make(map[string]struct{}, len(r.Functions))
			idx[r.Path] = m
		}
		for _, fn := range r.Functions {
			m[signatureKey(fn)] = struct{}{}
		}
	}
	return idx
}

func diffOnlyNew(newReports []model.FileReport, oldIndex map[string]map[string]struct{}) []model.FileReport {
	var out []model.FileReport
	for _, r := range newReports {
		oldSet := oldIndex[r.Path]
		var keep []model.FunctionDecl
		for _, fn := range r.Functions {
			key := signatureKey(fn)
			if oldSet == nil {
				keep = append(keep, fn)
				continue
			}
			if _, ok := oldSet[key]; !ok {
				keep = append(keep, fn)
			}
		}
		if len(keep) == 0 {
			continue
		}
		out = append(out, model.FileReport{Path: r.Path, Functions: keep})
	}
	return out
}

func signatureKey(fn model.FunctionDecl) string {
	// Exact match on normalized output: name + '(' + params + ')'
	return fn.Name + "(" + fn.Params + ")"
}

func (a *App) parseFiles(ctx context.Context, files []string, workers int) ([]model.FileReport, error) {
	if a.logger.Enabled(ctx, slog.LevelDebug) {
		a.logger.Debug("parsing started", "files", len(files), "workers", workers)
	}

	jobs := make(chan string, workers*2)
	results := make(chan model.FileReport, workers*2)
	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		workerID := w
		go func() {
			defer wg.Done()
			if a.logger.Enabled(ctx, slog.LevelDebug) {
				a.logger.Debug("worker started", "worker", workerID)
			}
			defer func() {
				if a.logger.Enabled(ctx, slog.LevelDebug) {
					a.logger.Debug("worker stopped", "worker", workerID)
				}
			}()
			for path := range jobs {
				if ctx.Err() != nil {
					return
				}
				var started time.Time
				if a.logger.Enabled(ctx, slog.LevelDebug) {
					started = time.Now()
					a.logger.Debug("parse started", "worker", workerID, "file", path)
				}
				report, err := a.parseOne(ctx, path)
				if err != nil {
					a.logger.Error("parse failed", "worker", workerID, "file", path, "err", err)
					select {
					case errCh <- err:
					default:
					}
					return
				}
				if a.logger.Enabled(ctx, slog.LevelDebug) {
					a.logger.Debug(
						"parse finished",
						"worker", workerID,
						"file", path,
						"functions", len(report.Functions),
						"duration_ms", time.Since(started).Milliseconds(),
					)
				}
				results <- report
			}
		}()
	}

	go func() {
		for _, f := range files {
			select {
			case <-ctx.Done():
				close(jobs)
				wg.Wait()
				close(results)
				return
			case jobs <- f:
				if a.logger.Enabled(ctx, slog.LevelDebug) {
					a.logger.Debug("job queued", "file", f)
				}
			}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	var out []model.FileReport
	for {
		select {
		case err := <-errCh:
			if err != nil {
				return nil, err
			}
		case r, ok := <-results:
			if !ok {
				return out, nil
			}
			out = append(out, r)
		}
	}
}

func (a *App) parseOne(ctx context.Context, path string) (model.FileReport, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return model.FileReport{}, fmt.Errorf("read %s: %w", path, err)
	}
	bytes := len(content)
	decls := a.parser.Parse(content)
	if a.logger.Enabled(ctx, slog.LevelDebug) {
		// Intentionally avoid logging the full signature list; file-level counts are sufficient.
		a.logger.Debug("file parsed", "file", path, "bytes", bytes, "functions", len(decls))
	}
	return model.FileReport{Path: path, Functions: decls}, nil
}

func (a *App) openOutput(ctx context.Context, reportPath string) (io.Writer, func(), error) {
	if reportPath == "" {
		if a.logger.Enabled(ctx, slog.LevelDebug) {
			a.logger.Debug("report output selected", "output", "stdout")
		}
		return os.Stdout, func() {}, nil
	}
	f, err := os.Create(reportPath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("create report file: %w", err)
	}
	if a.logger.Enabled(ctx, slog.LevelDebug) {
		a.logger.Debug("report output selected", "output", reportPath)
	}
	return f, func() { _ = f.Close() }, nil
}

func outputName(reportPath string) string {
	if reportPath == "" {
		return "stdout"
	}
	return reportPath
}

package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

type Config struct {
	Path      string
	Recursive bool
	Report    string
	LogLevel  string
	Workers   int

	// Diff mode: compare two existing reports (text format produced by this tool)
	// and output only functions that are present in DiffNew but absent in DiffOld.
	DiffOld string
	DiffNew string
}

func Parse(args []string) (Config, error) {
	fs := flag.NewFlagSet("phpfuncparser", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // we print our own errors

	var cfg Config
	fs.StringVar(&cfg.Path, "path", "", "Path to a PHP file or directory")
	fs.StringVar(&cfg.DiffOld, "diff-old", "", "Old report file to compare (enables diff mode)")
	fs.StringVar(&cfg.DiffNew, "diff-new", "", "New report file to compare (enables diff mode)")
	fs.BoolVar(&cfg.Recursive, "recursive", false, "Recursively scan subdirectories")
	fs.StringVar(&cfg.Report, "report", "", "Report output file (defaults to stdout)")
	fs.StringVar(&cfg.LogLevel, "log", "info", "Log level: debug|info|warn|error")
	fs.IntVar(&cfg.Workers, "workers", 0, "Number of parsing workers (0 = auto)")

	if err := fs.Parse(args); err != nil {
		printUsageError(os.Stderr, err)
		return Config{}, err
	}
	// Validate mode.
	if cfg.DiffOld != "" || cfg.DiffNew != "" {
		if cfg.DiffOld == "" || cfg.DiffNew == "" {
			err := errors.New("-diff-old and -diff-new must be provided together")
			printUsageError(os.Stderr, err)
			return Config{}, err
		}
		if cfg.Path != "" {
			err := errors.New("-path cannot be used with -diff-old/-diff-new")
			printUsageError(os.Stderr, err)
			return Config{}, err
		}
		return cfg, nil
	}

	if cfg.Path == "" {
		err := errors.New("-path is required")
		printUsageError(os.Stderr, err)
		return Config{}, err
	}
	return cfg, nil
}

func printUsageError(w io.Writer, err error) {
	fmt.Fprintln(w, "Error:", err)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  phpfuncparser -path <file|dir> [-recursive] [-report <file>] [-log <level>] [-workers <n>]")
	fmt.Fprintln(w, "  phpfuncparser -diff-old <old_report.txt> -diff-new <new_report.txt> [-report <diff_report.txt>] [-log <level>]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  phpfuncparser -path ./src -recursive -report functions.txt")
	fmt.Fprintln(w, "  phpfuncparser -path ./index.php")
	fmt.Fprintln(w, "  phpfuncparser -diff-old functions_prev.txt -diff-new functions_new.txt -report functions_added.txt")
}

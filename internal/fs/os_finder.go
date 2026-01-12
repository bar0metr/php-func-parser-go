package fs

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type OSFinder struct{}

func NewOSFinder() *OSFinder { return &OSFinder{} }

func (f *OSFinder) PHPFiles(ctx context.Context, path string, recursive bool) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if isPHPFile(path) {
			return []string{path}, nil
		}
		return nil, errors.New("path is not a directory or a .php file")
	}

	var files []string
	if recursive {
		err = filepath.WalkDir(path, func(p string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if d.IsDir() {
				return nil
			}
			if isPHPFile(p) {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if e.IsDir() {
				continue
			}
			p := filepath.Join(path, e.Name())
			if isPHPFile(p) {
				files = append(files, p)
			}
		}
	}

	sort.Strings(files)
	return files, nil
}

func isPHPFile(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".php")
}

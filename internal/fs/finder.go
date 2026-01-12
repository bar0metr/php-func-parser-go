package fs

import "context"

type Finder interface {
	PHPFiles(ctx context.Context, path string, recursive bool) ([]string, error)
}

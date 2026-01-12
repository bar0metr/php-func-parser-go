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
}

func Parse(args []string) (Config, error) {
	fs := flag.NewFlagSet("phpfuncparser", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // we print our own errors

	var cfg Config
	fs.StringVar(&cfg.Path, "path", "", "Path to a PHP file or directory")
	fs.BoolVar(&cfg.Recursive, "recursive", false, "Recursively scan subdirectories")
	fs.StringVar(&cfg.Report, "report", "", "Report output file (defaults to stdout)")
	fs.StringVar(&cfg.LogLevel, "log", "info", "Log level: debug|info|warn|error")
	fs.IntVar(&cfg.Workers, "workers", 0, "Number of parsing workers (0 = auto)")

	if err := fs.Parse(args); err != nil {
		printUsageError(os.Stderr, err)
		return Config{}, err
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
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  phpfuncparser -path ./src -recursive -report functions.txt")
	fmt.Fprintln(w, "  phpfuncparser -path ./index.php")
}

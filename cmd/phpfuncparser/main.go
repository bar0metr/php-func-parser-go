package main

import (
	"context"
	"fmt"
	"os"

	"php-func-parser-go/internal/app"
	"php-func-parser-go/internal/config"
	"php-func-parser-go/internal/logging"
	"php-func-parser-go/internal/version"
)

func main() {
	fmt.Printf("Php-func-parser-go version: %s\n", version.Version)
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		os.Exit(2)
	}

	logger := logging.New(cfg.LogLevel)
	application := app.New(cfg, logger)

	if err := application.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

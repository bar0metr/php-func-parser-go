# php-func-parser-go

Ultra-fast PHP function signature extractor written in Go.

## Build

```bash
go mod tidy
go test ./...
go build -o bin/phpfuncparser ./cmd/phpfuncparser
```

## Usage

Scan a directory recursively and write report to a file:

```bash
./bin/phpfuncparser -path ./src -recursive -report functions.txt
```

Scan a single file and print to stdout:

```bash
./bin/phpfuncparser -path ./index.php
```

## CLI Options

### `-path` (required)
Path to a **PHP file** or a **directory** to scan.

Examples:
```bash
./bin/phpfuncparser -path ./index.php
./bin/phpfuncparser -path ./src
```

### `-recursive` (optional)
If `-path` points to a directory, enables **recursive** traversal of subdirectories.

Default: `false`

Example:
```bash
./bin/phpfuncparser -path ./src -recursive
```

### `-report` (optional)
Output report destination.

- If set: writes to the specified file (created or truncated).
- If not set: prints to **stdout**.

Example:
```bash
./bin/phpfuncparser -path ./src -recursive -report functions.txt
```

### `-workers` (optional)
Number of concurrent workers used for scanning files.

- `0` means **auto** (runtime-based default).
- For SSD/NVMe, higher values may improve throughput; for HDD/network FS, too high may degrade performance.

Default: `0`

Examples:
```bash
./bin/phpfuncparser -path ./src -recursive -workers 0
./bin/phpfuncparser -path ./src -recursive -workers 16
```

### `-log` (optional)
Log verbosity level.

Supported values:
- `debug`
- `info`
- `warn`
- `error`

Default: `info`

Examples:
```bash
./bin/phpfuncparser -path ./src -recursive -log debug
./bin/phpfuncparser -path ./src -recursive -log error
```

## Output format

```
/path/to/file1.php
function foo($a, $b = 1)
function bar()

/path/to/file2.php
function baz(int $x)
```

## Notes

- Only **named** `function <name>(...)` declarations are reported (anonymous functions are skipped).
- The parser is a purpose-built single-pass scanner that ignores strings and comments.
- Logging goes to **stderr**; the report goes to `-report` or **stdout**.

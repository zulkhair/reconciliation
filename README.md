# Reconciliation App

Simple Reconciliation App in Go

## Project Structure

reconciliation/
├── cmd/
│ └── main.go # Main application entry point
├── pkg/
│ └── csv/ # CSV processing utilities
│ └── reconcile/ # Reconciliation logic
│ └── types/ # Shared types and constants
├── sample/ # Sample CSV files for testing
├── go.mod # Go module file
├── go.sum
├── README.md
└── Makefile

## Usage

Flags:
  -b, --bank string     Directory path contains bank statement CSV files or Comma-separated paths to bank statement CSV files (required)
  -e, --end string      End date for reconciliation in YYYY-MM-DD format (required)
  -h, --help            help for this command
  -o, --output string   Path to output JSON file
  -t, --start string    Start date for reconciliation in YYYY-MM-DD format (required)
  -s, --system string   Path to system transaction CSV file (required)

### Using go run command
```bash
go run cmd/main.go -s sample/matched/system.csv -b sample/matched/mandiri.csv -t 2024-01-01 -e 2024-01-31 -o output.json
```

### Using Makefile
```bash
make run system=sample/matched/system.csv bank=sample/matched/mandiri.csv start=2024-01-01 end=2024-01-31 output=output.json
```

## Build

### Using go build command

```bash
go build -o bin/reconciliation cmd/main.go
```

### Using Makefile

```bash
make build
```

### Run the binary after build
```bash
./bin/reconciliation -s sample/matched/system.csv -b sample/matched/mandiri.csv -t 2024-01-01 -e 2024-01-31 -o output.json
```


# Reconciliation App

Simple Reconciliation App in Go

## Input

- System Transaction CSV file
- Bank Statement CSV file
- Date range (start date and end date)

## Output

- JSON file (can be generated using flag --output)

## Project Structure

```
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
```

## Usage

```
Flags:
  -s, --system string   Path to system transaction CSV file (required)
  -b, --bank string     Directory path contains bank statement CSV files or Comma-separated paths to bank statement CSV files (required)
  -t, --start string    Start date for reconciliation in YYYY-MM-DD format (required)
  -e, --end string      End date for reconciliation in YYYY-MM-DD format (required)
  -o, --output string   Path to output JSON file
  -h, --help            help for this command
```

### Using go run command
```bash
# Example run using go run command
go run cmd/main.go -s sample/matched/system.csv -b sample/matched/mandiri.csv -t 2024-01-01 -e 2024-01-31 -o output.json
```

### Using Makefile
```bash
# makefile mask the input arguments
make run system=sample/matched/system.csv bank=sample/matched/mandiri.csv start=2024-01-01 end=2024-01-31 output=output.json
```

### Using Docker (if you don't have Go installed)

```bash
# Run with default arguments specified in docker-compose.yml
docker-compose up

# Run with custom arguments
docker-compose run --rm reconciliation \
  --system /app/data/multiple/system.csv \
  --bank /app/data/multiple/banks \
  --start 2024-01-01 \
  --end 2024-12-31 \
  --output /app/data/custom-result.json
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

## Note & Improvement (TODO)

- Tried to process 100.000 system transactions and 100.000 bank statements, it takes 2 minutes to be processed. (still slow)
- Need to try using database to store the data and use database query to get the data. (maybe faster)

# Reconciliation App

A Go-based financial reconciliation system that matches system transactions with bank statements from multiple banks. The application processes CSV files and identifies matched transactions, unmatched entries, and discrepancies.

## Features

- 📊 Multi-bank statement reconciliation
- 📁 CSV file processing for system and bank transactions
- 🔍 Automatic transaction matching by date and amount
- 📈 Discrepancy detection and reporting
- 🚀 RESTful API with Swagger documentation
- ✅ Comprehensive test coverage (84%+ service, 100% handler)

## Prerequisites

Before running the application, ensure you have the following installed:

- **Go** 1.21 or higher ([Download](https://golang.org/dl/))
- **Make** (usually pre-installed on macOS/Linux)
- **Swag** for API documentation generation
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```

## Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/elkoshar/reconciliation-app.git
   cd reconciliation-app
   ```

2. **Install dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Configure the application**
   
   Copy the sample configuration file:
   ```bash
   cp configs/sample-config.env configs/config.env
   ```
   
   Edit `configs/config.env` to customize settings:
   ```env
   SERVER_HTTP_PORT=8080
   LOG_LEVEL=INFO
   HTTP_READ_TIMEOUT=10s
   HTTP_WRITE_TIMEOUT=10s
   HTTP_INBOUND_TIMEOUT=10s
   HTTP_TIMEOUT=10s
   HTTP_DEBUG=true
   HTTP_MAX_IDLE_CONNECTIONS=100
   HTTP_MAX_IDLE_CONNECTIONS_PER_HOST=100
   HTTP_IDLE_CONNECTION_TIMEOUT=10s
   ```

## Running the Application

### Development Mode

```bash
# Build and run the HTTP server
make run-http
```

The server will start on `http://localhost:8080` by default.

## API Documentation

Once the application is running, access the Swagger documentation at:

```
http://localhost:8080/swagger/index.html
```

## Testing the Application

### API Testing with cURL

**Endpoint:** `POST /reconciliation-app/reconciliation`

**Basic Request:**

```bash
curl -X POST 'http://localhost:8080/reconciliation-app/reconciliation' \
  --header 'accept: application/json' \
  --header 'Accept-Language: id' \
  --form 'start_date=2025-11-01' \
  --form 'end_date=2025-11-30' \
  --form 'system_data=@"/path/to/System_Transactions.csv"' \
  --form 'bank_csv=@"/path/to/BCA_Statement.csv"' \
  --form 'bank_csv=@"/path/to/BRI_Statement.csv"' \
  --form 'bank_csv=@"/path/to/Mandiri_Statement.csv"'
```

**Example with actual file paths (using sample files from csv/ folder):**

```bash
curl -X POST 'http://localhost:8080/reconciliation-app/reconciliation' \
  --header 'accept: application/json' \
  --header 'Accept-Language: id' \
  --form 'start_date=2025-11-01' \
  --form 'end_date=2025-11-30' \
  --form 'system_data=@"csv/System_Transactions - Sheet1.csv"' \
  --form 'bank_csv=@"csv/BCA_Statement - Sheet1.csv"' \
  --form 'bank_csv=@"csv/BRI_Statement - Sheet1.csv"' \
  --form 'bank_csv=@"csv/Mandiri_Statement - Sheet1.csv"'
```

### CSV File Format

#### System Transactions CSV Format

```csv
trxID,amount,type,transactionTime
trx-bca-01,100000,DEBIT,2025-11-01 8:00:00
trx-bca-02,150000,DEBIT,2025-11-01 9:00:00
trx-bca-03,200000,DEBIT,2025-11-01 10:00:00
```

**Fields:**
- `trxID`: Unique transaction identifier
- `amount`: Transaction amount (positive number, in cents/smallest currency unit)
- `type`: Transaction type (`CREDIT` or `DEBIT`)
- `transactionTime`: Transaction timestamp (format: `YYYY-MM-DD H:MM:SS`)

**Sample file:** `csv/System_Transactions - Sheet1.csv`

#### Bank Statement CSV Format

```csv
unique_identifier,amount,date
bca-01,-100000,2025-11-01
bca-02,-150000,2025-11-01
bca-03,-200000,2025-11-01
```

**Fields:**
- `unique_identifier`: Unique bank transaction identifier
- `amount`: Transaction amount (negative for debits, positive for credits, in cents/smallest currency unit)
- `date`: Transaction date (format: `YYYY-MM-DD`)

**Sample files:** 
- `csv/BCA_Statement - Sheet1.csv`
- `csv/BRI_Statement - Sheet1.csv`
- `csv/Mandiri_Statement - Sheet1.csv`

### Expected Response

```json
{
    "data": {
        "TotalProcessed": 56,
        "TotalMatched": 22,
        "TotalUnmatched": 12,
        "TotalDiscrepancies": 5000.00,
        "UnmatchedSystem": [
            {
                "TransactionID": "trx-sys-only-01",
                "Amount": 11111.00,
                "Type": "DEBIT",
                "TransactionTime": "2025-11-20T08:00:00Z"
            },
            {
                "TransactionID": "trx-sys-only-02",
                "Amount": 22222.00,
                "Type": "CREDIT",
                "TransactionTime": "2025-11-21T08:00:00Z"
            },
            {
                "TransactionID": "trx-sys-only-03",
                "Amount": 33333.00,
                "Type": "DEBIT",
                "TransactionTime": "2025-11-22T08:00:00Z"
            }
        ],
        "UnmatchedBank": {
            "Stmt-BCA_Statement - Sheet1.csv": [
                {
                    "BankName": "Stmt-BCA_Statement - Sheet1.csv",
                    "UniqueID": "bca-08",
                    "Amount": -100.00,
                    "Date": "2025-11-01T00:00:00Z"
                },
                {
                    "BankName": "Stmt-BCA_Statement - Sheet1.csv",
                    "UniqueID": "bca-09",
                    "Amount": -200.00,
                    "Date": "2025-11-01T00:00:00Z"
                },
                {
                    "BankName": "Stmt-BCA_Statement - Sheet1.csv",
                    "UniqueID": "bca-10",
                    "Amount": 300.00,
                    "Date": "2025-11-01T00:00:00Z"
                }
            ],
            "Stmt-BRI_Statement - Sheet1.csv": [
                {
                    "BankName": "Stmt-BRI_Statement - Sheet1.csv",
                    "UniqueID": "bri-08",
                    "Amount": -500.00,
                    "Date": "2025-11-03T00:00:00Z"
                },
                {
                    "BankName": "Stmt-BRI_Statement - Sheet1.csv",
                    "UniqueID": "bri-09",
                    "Amount": -600.00,
                    "Date": "2025-11-03T00:00:00Z"
                },
                {
                    "BankName": "Stmt-BRI_Statement - Sheet1.csv",
                    "UniqueID": "bri-10",
                    "Amount": 700.00,
                    "Date": "2025-11-03T00:00:00Z"
                }
            ],
            "Stmt-Mandiri_Statement - Sheet1.csv": [
                {
                    "BankName": "Stmt-Mandiri_Statement - Sheet1.csv",
                    "UniqueID": "man-08",
                    "Amount": -10.00,
                    "Date": "2025-11-02T00:00:00Z"
                },
                {
                    "BankName": "Stmt-Mandiri_Statement - Sheet1.csv",
                    "UniqueID": "man-09",
                    "Amount": 20.00,
                    "Date": "2025-11-02T00:00:00Z"
                },
                {
                    "BankName": "Stmt-Mandiri_Statement - Sheet1.csv",
                    "UniqueID": "man-10",
                    "Amount": 30.00,
                    "Date": "2025-11-02T00:00:00Z"
                }
            ]
        }
    },
    "error": {
        "status": false,
        "msg": "",
        "code": 0
    },
    "message": "",
    "serverTime": 1763739105,
    "code": 200
}
```

## Project Structure

```
reconciliation-app/
├── api/
│   ├── http/
│   │   ├── reconciliation/      # HTTP handlers for reconciliation
│   │   ├── route.go             # Route definitions
│   │   └── server.go            # HTTP server configuration
│   ├── interface.go             # Service interfaces
│   └── middleware.go            # HTTP middlewares
├── bin/                         # Compiled binaries
├── cmd/
│   └── http/
│       └── main.go              # Application entry point
├── configs/
│   ├── config.go                # Configuration loader
│   ├── config.env               # Environment configuration
│   └── sample-config.env        # Sample configuration
├── docs/                        # Swagger documentation (auto-generated)
├── pkg/
│   ├── constants/               # Application constants
│   ├── helpers/                 # Helper functions
│   ├── logger/                  # Logging utilities
│   ├── panics/                  # Panic recovery
│   ├── response/                # HTTP response utilities
│   └── validator/               # Request validation
├── server/
│   ├── http.go                  # HTTP server implementation
│   └── server.go                # Server initialization
├── service/
│   └── reconciliation/          # Reconciliation business logic
│       ├── entity.go            # Data models
│       ├── service.go           # Service implementation
│       └── service_test.go      # Unit tests
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── Makefile                     # Build automation
└── README.md                    # This file
```

## Troubleshooting

### Common Issues

**1. Port already in use**
```bash
# Change the port in configs/config.env
SERVER_HTTP_PORT=8081
```

**2. Swagger not generating**
```bash
# Reinstall swag
go install github.com/swaggo/swag/cmd/swag@latest

# Regenerate docs
make swag
```

**3. Module errors**
```bash
# Clean and reinstall dependencies
go clean -modcache
go mod vendor
go mod tidy
```

**4. Test failures**
```bash
# Run tests with verbose output to see details
go test ./... -v
```

## Performance Considerations

- Maximum file size: 10MB per file
- Recommended batch size: Up to 10,000 transactions per file
- Concurrent bank file processing: Supported
- Memory usage scales with file size

## Contributors

@elkoshar

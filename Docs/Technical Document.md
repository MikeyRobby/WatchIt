# WatchIt — Technical Document

> **Version:** 0.1  
> **Last Updated:** April 2026  
> **Status:** In Progress

---
Check out the [[WatchIt/Docs/README|README]] for a breakdown of the project.
## 1. System Requirements

|Requirement|Minimum|
|---|---|
|**OS**|Windows 10, macOS 12, or Ubuntu 22.04|
|**Go**|1.22+|
|**Node.js**|18+ (for Wails frontend build)|
|**Wails CLI**|v2|

> **End users need none of the above.** The production build is a single self-contained executable. No installs, no configuration files, no database setup.

---

## 2. Project Structure

```
watchit/
├── cmd/
│   └── main.go                  # Wails entry point — bootstraps app and binds services
├── internal/
│   ├── app.go                   # Wails App struct — exposes Go methods to the frontend
│   ├── monitor/
│   │   ├── handler.go           # Wails-bound methods for the Monitor module
│   │   ├── service.go           # Business logic: search, polling, trend calculation
│   │   ├── repository.go        # sqlc-generated query wrappers for Monitor entities
│   │   └── ebay/
│   │       └── client.go        # eBay Browse + Marketplace Insights API client
│   ├── toolbox/
│   │   ├── handler.go           # Wails-bound methods for the Toolbox module
│   │   ├── service.go           # Business logic: work orders, invoices, client mgmt
│   │   ├── repository.go        # sqlc-generated query wrappers for Toolbox entities
│   │   └── pdf/
│   │       └── report.go        # PDF report generation using go-pdf/fpdf
│   ├── settings/
│   │   └── settings.go          # Persists user settings (eBay credentials, poll interval)
│   ├── db/
│   │   ├── db.go                # SQLite connection + auto-migration on first launch
│   │   ├── query/               # Raw .sql query files (input to sqlc)
│   │   └── sqlc/                # sqlc-generated Go code (do not edit manually)
│   └── logger/
│       └── logger.go            # Structured logging setup
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── monitor/         # Monitor-specific UI components
│   │   │   └── toolbox/         # Toolbox-specific UI components
│   │   ├── pages/
│   │   │   ├── Setup.jsx        # First-launch eBay credentials screen
│   │   │   ├── Dashboard.jsx
│   │   │   ├── Watchlist.jsx
│   │   │   ├── Collections.jsx
│   │   │   ├── WorkOrders.jsx
│   │   │   ├── Clients.jsx
│   │   │   ├── Invoices.jsx
│   │   │   └── Settings.jsx
│   │   └── main.jsx
│   ├── package.json
│   └── vite.config.js
├── docs/
│   ├── design.md                # Design document
│   ├── technical.md             # This document
│   └── diagrams/
│       └── architecture.png
├── Makefile
└── wails.json
```

---

## 3. Setup & Running Locally

### 3.1 Clone the Repository

```bash
git clone https://github.com/yourname/watchit.git
cd watchit
```

### 3.2 Install Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 3.3 Run in Development Mode

```bash
wails dev
```

This starts the app with hot-reload on both the Go backend and React frontend. The SQLite database is created automatically at `~/.watchit/watchit.db` on first run.

### 3.4 Build a Production Executable

```bash
wails build
```

Output is placed in `build/bin/`. This is the single file end users download and run.

---

## 4. First-Launch User Experience

End users have no setup steps beyond downloading and opening the executable. On first launch, WatchIt:

1. Creates a SQLite database file at the OS-appropriate user data directory (`%APPDATA%\WatchIt\` on Windows, `~/Library/Application Support/WatchIt/` on macOS, `~/.watchit/` on Linux)
2. Runs all schema migrations automatically against the new database
3. Detects that no eBay credentials are saved and presents the **Setup Screen** — a simple form asking for the user's eBay Client ID and Client Secret
4. Saves credentials to the same user data directory and proceeds to the main app

eBay credentials can be updated at any time via the Settings page. The Toolbox works fully offline with no credentials required.

---

## 5. Configuration & Settings

There is no `.env` file for end users. All user-facing configuration is stored in a `settings.json` file in the user data directory and managed through the in-app Settings page.

|Setting|Description|Default|
|---|---|---|
|`ebay_client_id`|eBay Developer app Client ID|_(set on first launch)_|
|`ebay_client_secret`|eBay Developer app Client Secret|_(set on first launch)_|
|`ebay_env`|`PRODUCTION` or `SANDBOX`|`PRODUCTION`|
|`poll_interval_minutes`|How often to poll eBay for new prices|`60`|
|`db_path`|Path to the SQLite database file|_(OS default)_|

For developers, `EBAY_ENV` can be overridden to `SANDBOX` via an environment variable at run time to test against the eBay sandbox without touching user settings.

---

## 6. Database Schema

The schema is applied automatically on first launch via embedded SQL migrations in `internal/db/db.go`. There is no manual migration step for end users.

SQLite does not support `UUID` or `TIMESTAMPTZ` natively — these are stored as `TEXT` (UUID strings) and `TEXT` (ISO 8601 timestamps) respectively.

### Monitor Tables

```sql
-- Canonical watch identity (brand + model + reference)
CREATE TABLE IF NOT EXISTS watches (
    id               TEXT PRIMARY KEY,
    brand            TEXT NOT NULL,
    model            TEXT NOT NULL,
    reference_number TEXT,
    created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- Watches the user wants to buy (not yet owned)
CREATE TABLE IF NOT EXISTS watchlist_entries (
    id           TEXT PRIMARY KEY,
    watch_id     TEXT NOT NULL REFERENCES watches(id) ON DELETE CASCADE,
    target_price REAL,
    notes        TEXT,
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- User-defined groupings of owned watches
CREATE TABLE IF NOT EXISTS collections (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS collection_items (
    id             TEXT PRIMARY KEY,
    collection_id  TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    watch_id       TEXT NOT NULL REFERENCES watches(id) ON DELETE CASCADE,
    purchase_price REAL,
    purchase_date  TEXT,
    notes          TEXT
);

-- Time-series price records pulled from eBay
CREATE TABLE IF NOT EXISTS price_records (
    id          TEXT PRIMARY KEY,
    watch_id    TEXT NOT NULL REFERENCES watches(id) ON DELETE CASCADE,
    source      TEXT NOT NULL,   -- 'ebay_active' | 'ebay_sold'
    price       REAL NOT NULL,
    currency    TEXT NOT NULL DEFAULT 'USD',
    listing_url TEXT,
    recorded_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_price_records_watch_recorded
    ON price_records(watch_id, recorded_at DESC);
```

### Toolbox Tables

```sql
CREATE TABLE IF NOT EXISTS clients (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    email      TEXT,
    phone      TEXT,
    notes      TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS work_orders (
    id                TEXT PRIMARY KEY,
    client_id         TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    watch_description TEXT NOT NULL,
    scope_of_work     TEXT NOT NULL,
    status            TEXT NOT NULL DEFAULT 'open',  -- open | in_progress | complete
    created_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    completed_at      TEXT
);

CREATE TABLE IF NOT EXISTS work_order_photos (
    id            TEXT PRIMARY KEY,
    work_order_id TEXT NOT NULL REFERENCES work_orders(id) ON DELETE CASCADE,
    file_path     TEXT NOT NULL,
    caption       TEXT,
    taken_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS invoices (
    id            TEXT PRIMARY KEY,
    work_order_id TEXT NOT NULL REFERENCES work_orders(id) ON DELETE CASCADE,
    line_items    TEXT NOT NULL DEFAULT '[]',  -- JSON: [{description, quantity, unit_price}]
    total         REAL NOT NULL,
    issued_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    paid_at       TEXT
);

CREATE TABLE IF NOT EXISTS reports (
    id            TEXT PRIMARY KEY,
    work_order_id TEXT NOT NULL REFERENCES work_orders(id),
    invoice_id    TEXT REFERENCES invoices(id),
    pdf_path      TEXT NOT NULL,
    generated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
```

---

## 6. Wails Binding Reference

Wails exposes Go methods to the frontend as callable JavaScript functions via `window.go.<Method>()`. All bound methods live on the `App` struct in `internal/app.go`.

### Monitor Bindings

```go
// Search for a watch on eBay, returns current price data
func (a *App) SearchWatch(brand, model, reference string) ([]PriceResult, error)

// Add a watch to the user's watchlist
func (a *App) AddToWatchlist(watchID string, targetPrice float64, notes string) error

// Get all watchlist entries with their latest price
func (a *App) GetWatchlist() ([]WatchlistEntry, error)

// Get price history for a watch (for trend graph)
func (a *App) GetPriceHistory(watchID string, days int) ([]PriceRecord, error)

// Create a new collection
func (a *App) CreateCollection(name, description string) (Collection, error)

// Add a watch to a collection
func (a *App) AddToCollection(collectionID, watchID string, purchasePrice float64) error

// Get all collections with total estimated values
func (a *App) GetCollections() ([]CollectionSummary, error)
```

### Toolbox Bindings

```go
// Client management
func (a *App) CreateClient(name, email, phone, notes string) (Client, error)
func (a *App) GetClient(clientID string) (ClientProfile, error)
func (a *App) ListClients() ([]Client, error)

// Work orders
func (a *App) CreateWorkOrder(clientID, watchDescription, scopeOfWork string) (WorkOrder, error)
func (a *App) UpdateWorkOrderStatus(workOrderID, status string) error
func (a *App) AttachPhoto(workOrderID, filePath, caption string) error
func (a *App) GetWorkOrder(workOrderID string) (WorkOrderDetail, error)

// Invoices
func (a *App) CreateInvoice(workOrderID string, lineItems []LineItem) (Invoice, error)
func (a *App) MarkInvoicePaid(invoiceID string) error

// Reports
func (a *App) GenerateReport(workOrderID string) (string, error) // returns output PDF path
```

### Settings Bindings

```go
// Returns true if eBay credentials have been saved (used to trigger setup screen)
func (a *App) IsConfigured() bool

// Save eBay credentials entered on setup screen or settings page
func (a *App) SaveEbayCredentials(clientID, clientSecret, env string) error

// Get current settings for the settings page UI
func (a *App) GetSettings() (Settings, error)

// Update poll interval
func (a *App) SetPollInterval(minutes int) error
```

Prices are kept current by a background goroutine that runs on startup and polls on a configurable interval.

```
App Start
    │
    ▼
IsConfigured? ──No──► Show Setup Screen
    │
   Yes
    │
    ▼
Load all watches from DB
    │
    ▼
┌─── Ticker (every N minutes) ───────────────────────────┐
│                                                         │
│   For each watch:                                       │
│       ├── Fetch eBay active listings (Browse API)      │
│       ├── Fetch eBay sold listings (Insights API)      │
│       └── Write PriceRecord rows to SQLite             │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**eBay OAuth flow:** The eBay client uses Client Credentials grant (app-level token, no user login required). Tokens are cached in memory and refreshed automatically when expired.

---

## 8. Toolbox: PDF Report Generation

Reports are generated using `go-pdf/fpdf` and saved to a local `reports/` directory.

Each report includes:

1. **Cover section** — client name, date, watch description
2. **Scope of work** — from the work order
3. **Photo gallery** — each attached photo with its caption, laid out in sequence
4. **Invoice summary** — line items and total
5. **Footer** — generated timestamp, report ID for reference

```go
// Simplified generation flow in internal/toolbox/pdf/report.go
func GenerateReport(order WorkOrderDetail, invoice Invoice, outputPath string) error {
    pdf := fpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    writeCoverSection(pdf, order)
    writeScopeSection(pdf, order)
    writePhotoGallery(pdf, order.Photos)
    writeInvoiceSection(pdf, invoice)
    writeFooter(pdf)
    return pdf.OutputFileAndClose(outputPath)
}
```

---

## 9. External Dependencies

|Package|Purpose|
|---|---|
|`github.com/wailsapp/wails/v2`|Desktop app framework|
|`modernc.org/sqlite`|Pure Go SQLite driver (no CGO, bundles cleanly)|
|`github.com/sqlc-dev/sqlc`|Type-safe SQL code generation|
|`github.com/go-pdf/fpdf`|PDF report generation|

---

## 10. Makefile Reference

```makefile
make run          # Start the app in Wails dev mode (hot reload)
make build        # Build production executable via Wails
make sqlc         # Regenerate sqlc Go code from query files
make test         # Run all Go tests
make lint         # Run golangci-lint
```

---

## 11. Known Limitations

- **PDF photo layout** is fixed-grid and does not adapt to portrait vs landscape photos elegantly.
- **No data backup mechanism** — users are responsible for backing up their `.db` file. The file location is shown in the Settings page to make this easier.
- **eBay Sandbox vs Production** — the app defaults to Production. Developers can override to Sandbox via the `EBAY_ENV` environment variable at run time without modifying user settings.
- **Price polling is per-watch** — with a large watchlist this could approach eBay API rate limits. Batching is not yet implemented.
- **SQLite concurrent writes** — SQLite handles one write at a time. The background polling goroutine and any user actions that write simultaneously are serialized automatically, but under heavy use this could cause brief UI delays. This is acceptable at the current scale.
- **No schema rollback** — migrations are applied automatically on launch and there is no rollback mechanism for end users. Breaking schema changes require careful migration planning.

Check out the [[WatchIt/Docs/README|README]] for the project overview and feature list


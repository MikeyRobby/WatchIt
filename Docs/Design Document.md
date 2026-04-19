Check out the [[WatchIt/Docs/README|README]] for a project overview.
# WatchIt — Design Document

> **Version:** 0.1  
> **Last Updated:** April 2026  
> **Status:** In Progress

---

## 1. Problem Statement

Watch enthusiasts, independent repair technicians, and small watch shops currently have no unified tool designed for their specific needs. Hobbyists managing a collection must manually track values across online marketplaces like eBay, while repair shops rely on generic invoicing software that lacks any concept of watch servicing, work order photo documentation, or client history.

WatchIt solves both problems in a single desktop application — a **Monitor** for tracking watch prices and collection value, and a **Toolbox** for managing repair work, clients, and professional reporting.

---

## 2. Target Users

|User|Description|
|---|---|
|**Collector**|Enthusiast with a personal collection who wants to track market value and monitor potential buys|
|**Hobbyist Repairer**|Individual who repairs watches as a side business and needs lightweight job tracking|
|**Independent Shop**|Small watch repair shop needing client management, work orders, invoicing, and client-facing reports|

---

## 3. Goals

- Allow users to monitor watch prices across eBay over time with trend visibility
- Allow users to create and manage watchlists for potential purchases and collections for owned pieces
- Allow shops and hobbyists to create work orders with step-by-step photo documentation
- Allow generation of professional PDF reports that can be shared with clients
- Distribute as a single downloadable executable — no browser, no server, no technical setup required
- Get a non-technical user from download to running in under two minutes

---

## 4. Non-Goals

The following are explicitly out of scope for the initial version:

- **Mobile support** — WatchIt is a desktop-only application
- **Cloud sync or remote access** — all data is stored locally
- **Multi-user / team accounts** — single user per installation
- **Automated buying or selling** — price tracking is read-only, no marketplace integration beyond data retrieval
- **Support for non-watch items** — the domain model is watch-specific

---

## 5. User Stories

### Monitor

- As a collector, I want to search for a specific watch model so that I can add it to my watchlist and track its price over time.
- As a collector, I want to see a price trend graph for a watch so that I can identify good buying opportunities.
- As a collector, I want to create a collection of watches I own so that I can see the total estimated value of my collection.
- As a collector, I want a dashboard overview so that I can see all of my watchlist and collection pieces and their current spot prices at a glance.

### Toolbox

- As a repairer, I want to create a work order for a client's watch so that I have a formal record of the agreed scope of work.
- As a repairer, I want to attach photos with descriptions to a work order so that I can document each step of the service process.
- As a repairer, I want to generate an invoice tied to a work order so that I can bill my client accurately.
- As a repairer, I want to generate a PDF report of completed work so that I can share a professional summary with my client and keep it for my records.
- As a repairer, I want to view a client's profile and history so that I can quickly reference all previous work done for that client.

---

## 6. Architecture Overview

WatchIt is a desktop application built with **Wails**, which packages a Go backend and a web-based frontend into a single native executable. There is no HTTP server exposed to a network — the frontend communicates directly with Go via Wails' binding system.

```
┌─────────────────────────────────────────────┐
│                  WatchIt App                │
│                                             │
│  ┌─────────────┐       ┌─────────────────┐  │
│  │  Frontend   │◄─────►│   Go Backend    │  │
│  │  (React)    │       │                 │  │
│  │             │       │  ┌───────────┐  │  │
│  │  Monitor UI │       │  │  Monitor  │  │  │
│  │  Toolbox UI │       │  │  Service  │  │  │
│  └─────────────┘       │  └───────────┘  │  │
│                        │  ┌───────────┐  │  │
│                        │  │  Toolbox  │  │  │
│                        │  │  Service  │  │  │
│                        │  └───────────┘  │  │
│                        │  ┌───────────┐  │  │
│                        │  │  PDF Gen  │  │  │
│                        │  └───────────┘  │  │
│                        └────────┬────────┘  │
│                                 │           │
│                    ┌────────────▼─────────┐ │
│                    │  SQLite (.db file)   │ │
│                    │  auto-created on     │ │
│                    │  first launch        │ │
│                    └──────────────────────┘ │
└─────────────────────────────────────────────┘

                    ┌────────────────────────┐
                    │   eBay Developer API   │
                    │   Browse API           │
                    │   Marketplace Insights │
                    └────────────────────────┘
```

---

## 7. Key Design Decisions

### 7.1 Desktop App via Wails over a Web App

**Decision:** Build as a native desktop application using Wails rather than a web application.

**Considered:**

- Traditional web app (Go backend + browser frontend)
- Wails (Go backend + native desktop wrapper)
- Fyne (pure Go UI toolkit)

**Reasoning:** The target users are individuals and small shops — not teams sharing a remote system. A downloadable executable is a far lower barrier to adoption than asking someone to install Go, run a server, and open a browser. Wails was chosen over Fyne because it allows richer UI components (charts, image galleries for work order photos) using established web frameworks like React, while still keeping the entire backend in Go with no bridging overhead.

---

### 7.2 Go as the Backend Language

**Decision:** Use Go for all backend logic.

**Considered:**

- Python (fast to prototype, large ecosystem)
- Node.js (JavaScript across the stack)
- Go (strong concurrency, single binary output)

**Reasoning:** Go's compilation to a single static binary is a direct enabler of the desktop distribution model. Its concurrency model is well-suited to the background price polling that Monitor requires. The strong type system reduces runtime errors in what will eventually be financial/business data handling.

---

### 7.3 SQLite over PostgreSQL

**Decision:** Use SQLite as the local database rather than PostgreSQL.

**Considered:**

- PostgreSQL (more powerful querying, better time-series support)
- SQLite (zero setup, single file, bundles into the executable)

**Reasoning:** The primary design goal is that a non-technical user can download and run WatchIt with no setup. PostgreSQL requires a separate installation, a running service, and manual configuration — a hard barrier for the target audience. SQLite requires nothing: the database is a single `.db` file that is automatically created on first launch. For the scale of a personal collection or small shop, SQLite's query capabilities are entirely sufficient for price trend data. If a future version targets teams or larger shops, migrating to PostgreSQL is a viable path. `modernc.org/sqlite` is used specifically because it is pure Go with no CGO dependency, which ensures clean cross-platform Wails builds.

---

### 7.4 sqlc over an ORM

**Decision:** Use sqlc for database access rather than an ORM like GORM.

**Considered:**

- GORM (most popular Go ORM, active ecosystem)
- sqlc (SQL-first, generates type-safe Go code)
- Raw `database/sql`

**Reasoning:** sqlc lets you write plain SQL and generates fully type-safe Go query functions from it. For a project where the data model is well-understood upfront and raw query control matters (especially for price trend queries), sqlc gives the safety of an ORM without the magic or performance overhead. Raw `database/sql` was ruled out due to the verbosity and boilerplate involved in scanning rows.

---

### 7.5 eBay API as the Sole Price Source

**Decision:** Use the official eBay Developer API as the only price data source, dropping Chrono24.

**Considered:**

- eBay API + Chrono24 scraping (colly)
- eBay API only

**Reasoning:** Chrono24 has no public API, meaning any integration would rely on web scraping. Scrapers are brittle — a Chrono24 page layout change silently breaks price data with no error the user would understand. The eBay Developer platform provides two purpose-built APIs: the Browse API for active listings and the Marketplace Insights API for sold listing data. Sold prices are the more meaningful valuation signal. Dropping Chrono24 eliminates a maintenance liability, keeps the codebase simpler, and ensures price data reliability. Additional marketplaces with official APIs can be added in future versions.

---

### 7.6 go-pdf/fpdf for Report Generation

**Decision:** Use go-pdf/fpdf for generating client-facing PDF reports.

**Considered:**

- go-pdf/fpdf (pure Go, no CGO)
- wkhtmltopdf (HTML to PDF, requires external binary)
- Headless Chrome (renders HTML to PDF)

**Reasoning:** fpdf is a pure Go library with no external binary dependencies, which is critical for a distributable desktop app. wkhtmltopdf and headless Chrome both require bundling or assuming the presence of external executables, which complicates distribution significantly.

---

## 8. Data Model (Conceptual)

### Monitor

|Entity|Key Fields|
|---|---|
|**Watch**|`id`, `brand`, `model`, `reference_number`|
|**WatchlistEntry**|`id`, `watch_id`, `target_buy_price`, `notes`|
|**Collection**|`id`, `name`, `description`|
|**CollectionItem**|`id`, `collection_id`, `watch_id`, `purchase_price`, `purchase_date`|
|**PriceRecord**|`id`, `watch_id`, `source`, `price`, `currency`, `recorded_at`|

### Toolbox

|Entity|Key Fields|
|---|---|
|**Client**|`id`, `name`, `email`, `phone`, `notes`|
|**WorkOrder**|`id`, `client_id`, `watch_description`, `scope_of_work`, `status`, `created_at`|
|**WorkOrderPhoto**|`id`, `work_order_id`, `file_path`, `caption`, `taken_at`|
|**Invoice**|`id`, `work_order_id`, `line_items`, `total`, `issued_at`, `paid_at`|
|**Report**|`id`, `work_order_id`, `invoice_id`, `pdf_path`, `generated_at`|

---

## 9. Open Questions

- Should PDF reports be stored on disk or regenerated each time? Storing saves time but adds file management complexity.
- Should the eBay API polling interval be configurable by the user in settings, or fixed?
- Should eBay API credentials be entered through a first-launch setup screen, or through a persistent settings page accessible at any time? A setup screen is friendlier for new users; a settings page is more flexible long-term.

---

## 10. Future Considerations

- Cloud sync option for collectors who want access across devices
- Mobile companion app (read-only dashboard view)
- Support for additional marketplaces with official APIs (Chrono24 if they release one, Bob's Watches, Watchfinder)
- Multi-currency support for international users
- Migration path to PostgreSQL for shops that outgrow SQLite
# POS Phoenix — Pardis Barbershop Management & Point of Sale

A production-grade, mobile-first cash-flow POS and multi-branch backoffice management system engineered with Go, SQLite, server-rendered HTML templates, HTMX, Alpine.js, Tailwind CSS, and Playwright automated end-to-end testing. Designed specifically for daily operations at Pardis Barbershop, targeting mobile handheld terminals (Android 10 / Chrome 83) as well as desktop and tablet backoffice environments.

---

## 🌟 Key Features

1. **Multi-Branch Operations & Access Control**:
   - Native support for multiple branch locations (**Klaseman** and **Ledok**).
   - Role-based authorization hierarchy: `superadmin` (Owner), `manager`, `cashier`, and `barberman` (Operators).
   - Branch-scoped transaction filtering and administrative oversight.

2. **Exclusive Profit Sharing System (Owner / Ipang)**:
   - Dynamic monthly profit-sharing configuration per branch.
   - Dedicated splits for Owner (Ipang), individual staff/barbers, and an automated unallocated cash reserve balance.
   - Dual-layer validation (real-time client-side calculation and strict server-side validation enforcing total allocation $\le$ 100.0%).

3. **24-Month Financial Report Consolidation (Excel `.xlsx`)**:
   - Comprehensive multi-branch financial exports powered by `excelize`.
   - Multi-sheet workbook format:
     * `Consolidated Summary`: Company-wide aggregated revenue, product sales, profit share, operational expenses, and net profit.
     * `Branch Klaseman`: Granular service and retail performance for the Klaseman branch.
     * `Branch Ledok`: Granular financial records for the Ledok branch.
   - Professional spreadsheet typography, styled financial headers, and standard accounting number formats.

4. **Employee Payroll Slip Generator**:
   - Official Pardis Barbershop payroll slips featuring clean company branding, branch metadata, service commission breakdown, retail product commission details, and final take-home pay.
   - Interactive HTML preview with direct **"Print / Save PDF"** capability.
   - Bulk branch export option packaged as compressed ZIP archives.

5. **Point of Sale & Cash Flow Management**:
   - Fast daily transaction entry with integer-cent monetary precision (`int64`).
   - Active discount rules and product bundling support.
   - Thermal receipt printer formatting and reprint capability.
   - Bilingual localization support (Indonesian `id` and English `en`) and light/dark theme persistence.
   - Jakarta business-day operational boundaries (`Asia/Jakarta` / WIB) with UTC persistence.

---

## 📖 Backoffice API Documentation

All Backoffice endpoints reside under the `/backoffice` namespace and require an authenticated session with `superadmin` privileges.

### 1. 24-Month Financial Trend Analytics (JSON API)
* **Endpoint**: `GET /backoffice/api/analytics/trend-24m`
* **Description**: Returns historical financial metrics across the trailing 24 months for analytics dashboards and visualization charts.
* **Query Parameters**:
  * `branch` *(optional)*: Branch code filter (`KLASEMAN`, `LEDOK`, or omit for consolidated multi-branch aggregates).
* **Sample Response**:
  ```json
  {
    "period_from": "2024-10",
    "period_to": "2026-09",
    "months": [
      {
        "period": "2026-09",
        "service_revenue_cents": 1500000000,
        "product_sales_cents": 250000000,
        "gross_revenue_cents": 1750000000,
        "owner_share_cents": 300000000,
        "employee_share_cents": 900000000,
        "net_profit_cents": 550000000,
        "tx_count": 420
      }
    ]
  }
  ```

### 2. 24-Month Financial Excel Export
* **Endpoint**: `GET /backoffice/reports.xlsx`
* **Description**: Generates and downloads a multi-sheet OpenXML workbook (`.xlsx`) containing full financial history over the last 24 months.
* **Workbook Structure**:
  * `Ringkasan Konsolidasi`: Company-wide aggregated ledger.
  * `Cabang Klaseman`: Detailed service and product performance for Klaseman.
  * `Cabang Ledok`: Detailed service and product performance for Ledok.
* **HTTP Headers**:
  ```http
  Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
  Content-Disposition: attachment; filename="laporan-keuangan-2024-10-sd-2026-09.xlsx"
  ```

### 3. Profit Sharing Configuration
* **Endpoint**: `POST /backoffice/profit-sharing`
* **Description**: Updates monthly revenue-sharing percentages. Restricted strictly to Owner / Superadmin (`ipang`).
* **Content-Type**: `application/x-www-form-urlencoded`
* **Payload Fields**:
  * `csrf`: Valid CSRF session token.
  * `branch`: Target branch code (`KLASEMAN` or `LEDOK`).
  * `period`: Target month formatted as `YYYY-MM` (e.g., `2026-09`).
  * `owner_percentage`: Float representing owner share (0.0 – 100.0).
  * `employee_id`: Array of employee user IDs.
  * `employee_percentage`: Array of floats corresponding to each `employee_id`.
* **Validation Rules**:
  * Total `owner_percentage` + $\sum(\text{employee\_percentage}) \le 100.0\%$.
  * Remaining unallocated balance is automatically routed to the branch reserve fund.

### 4. Employee Payroll Slips & Bulk Archive
* **Individual Slip (HTML / PDF)**: `GET /backoffice/payroll/slip?branch=KLASEMAN&period=2026-09&employee_id=1`
  * Renders a printable official payroll slip.
* **Bulk Export (ZIP)**: `GET /backoffice/payroll/slip-all?branch=KLASEMAN&period=2026-09&format=zip`
  * Streams a compressed ZIP file containing payroll records for all branch employees.

---

## 🧪 Automated Testing Suite

### Running Go Unit Tests
Executes unit tests covering financial formulas, profit sharing bounds, database operations, and authorization rules:
```bash
go test -v ./internal/httpserver/... -run "TestBackoffice"
```
To run the entire project test suite:
```bash
go test ./...
```

### Running Automated E2E Tests (Playwright Headless)
Executes automated browser verification in headless mode across 3 core operational workflows (Owner Login, Profit Sharing Updates, 24-Month Excel Download, and Payroll Preview):
```bash
npm run test:e2e
```

---

## 🚀 Deployment Guide

### 1. Docker & Docker Compose (Recommended Production Setup)
POS Phoenix includes an optimized multi-stage [Dockerfile](file:///c:/Users/USER/Downloads/Yoga/Yogi/pos-phoenix/Dockerfile) and [compose.yaml](file:///c:/Users/USER/Downloads/Yoga/Yogi/pos-phoenix/compose.yaml).

```bash
# 1. Prepare production environment variables
cp .env.example .env

# 2. Build and launch containers in background
docker compose up -d --build

# 3. View live server logs
docker compose logs -f pos
```
The application will listen on port `8080`, with SQLite database state persisted safely within the `pos_data` Docker volume.

### 2. Native Deployment (Local Development / Windows / Linux)
Prerequisites: Go 1.22+ and Node.js 18+.

```bash
# 1. Install Node dependencies
npm install

# 2. Compile Tailwind CSS production bundle
npm run css:build

# 3. Launch HTTP server
go run ./cmd/server

# Alternatively on Windows Command Prompt / PowerShell:
.\run.bat
```
Navigate to `http://localhost:8080` in your web browser.

---

## ⚙️ Environment Variables Configuration

The server binary is configured through standard environment variables:

| Variable | Default | Description |
| :--- | :--- | :--- |
| `ADDR` | `:8080` | Network address and port for the HTTP listener (e.g., `:8080` or `:3000`). |
| `DATABASE_PATH` | `pos.db` | Filepath to the local SQLite database. |
| `BOOTSTRAP_ADMIN_EMAIL` | `admin@example.com` | Email address for initial superadmin bootstrapping. |
| `BOOTSTRAP_ADMIN_PASSWORD`| *(required in prod)* | Initial password for the bootstrap superadmin account. |
| `SESSION_SECURE` | `false` | Must be set to `true` when deployed behind HTTPS to enforce secure cookies. |
| `BUSINESS_TIMEZONE` | `Asia/Jakarta` | Local timezone for business-day boundaries and timestamps (WIB). |

---

## 👥 Seed & Development Accounts

* **Superadmin / Owner (Ipang)**:
  * Username / Email: `ipang` / `ipang@example.com`
  * Password: `adminsupervisor`
  * Scope: Global multi-branch oversight & profit-sharing configuration.
* **Operator / Barber (Yogi)**:
  * Username / Email: `yogi` / `yogi@contoh.com`
  * Password: `yogioperator`
  * Scope: Front-of-house cash register, order entry, and daily transactions.

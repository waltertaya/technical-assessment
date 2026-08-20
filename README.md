# Doctor's Appointment Booking System

A production-grade clinic booking backend built in **Go** and **PostgreSQL**, designed for a 5-doctor clinic that needs to scale. The system prevents double-booking under concurrent load, models real-world shift structures (including midday breaks), and calculates slot availability dynamically rather than pre-generating rows.

> Built for the Savannah Informatics Backend Developer Take-Home Assessment.

#### Submission
- [Public URL of Deployed application](http://16.170.143.11/api/v1/health)

---

## Table of Contents

- [System Design](#system-design)
- [Architecture Trade-offs](#architecture-above-trade-offs)
- [Local Setup](#local-setup)
- [API Routes](#api-routes)
- [Testing](#run-tests)
- [Deployment & CI/CD](#deployment--cicd)
- [AI Reflection](#ai-reflection)
- [Author](#author)

---

## System Design

The architecture is designed to be highly maintainable, prevent booking conflicts (race conditions), and support real-world clinic constraints such as irregular shift structures and midday breaks.

### Entity Relationship Diagram (ERD)

To represent a realistic clinic setup without over-complicating storage requirements, the system distinguishes between standard working-hour frames (shifts) and actual bookings — slots are never stored, only generated on the fly.

- **Working Hours (Shift)** — active, non-overlapping daily working shifts for a doctor on specific days of the week. Allowing multiple rows per day of the week means midday breaks are handled natively, with no special-case logic.
- **Appointment** — a 30-minute booking reservation linking a `Patient` and a `Doctor` on a specific calendar date.

[![ERD](</src/images/savannah informatics OA - doctor appointment.png>)](https://dbdiagram.io/d/savannah-informatics-OA-doctor-appointment-6a85bacd6440800f52a25235)

### Dynamic Availability Calculation Logic

Instead of pre-generating empty time slots in the database (which creates data-synchronization overhead and bloat as the clinic grows), the system **calculates slot availability dynamically on-the-fly** — cross-referencing a doctor's working-hour shifts against existing appointments for the requested date.

[![Availability logic](/src/images/image.png)](https://lucid.app/lucidchart/1be14c71-1c7a-4cf4-992c-610fe8ce0024/edit?viewport_loc=110%2C-2343%2C1950%2C2396%2C0_0&invitationId=inv_d00e573a-96e4-4b5c-9b1a-99b1afebc5e3)

### Concurrency Handling & Race Conditions

In high-traffic systems, multiple users might attempt to book the exact same 30-minute slot simultaneously. If untreated, this leads to a **double-booking bug**.

**The solution:** `SELECT ... FOR UPDATE` row-level locking on the doctor's record, inside a serializable transaction boundary, for every write path (`book`, `cancel`, `reschedule`). This forces concurrent requests against the same doctor to serialize at the database level, closing the race window entirely — without needing an external lock manager like Redis Redlock.

### Architecture (above) Trade-offs

| Design Dimension | Selected Approach | Alternatives Considered | Trade-off Analysis |
| :--- | :---: | :---: | :--- |
| Slot Generation | Dynamic Calculation | Static Pre-generation | **Pros:** Immediate adjustment of doctor working hours without retroactively editing rows. Zero row bloat as the clinic grows.<br>**Cons:** Marginally higher CPU load per query (mitigable with Redis caching at scale). |
| Concurrency Control | Database Lock (`FOR UPDATE`) | Go-Mutex / Distributed Locks | **Pros:** Robust and simple. Consistent across multiple backend instances without Redis Redlock orchestration.<br>**Cons:** Holds open transactional resources; mitigated by scoping locks strictly to the individual doctor. |
| Data Integrity | DB-level Constraints | Application-level validation only | **Pros:** Guarantees absolute data integrity even if another microservice or manual query touches the database directly.<br>**Cons:** Tighter coupling between business rules and the database schema. |

---

## Local Setup

### Prerequisites

- Go 1.21+
- PostgreSQL 15+

### Clone the repository

```bash
git clone git@github.com:waltertaya/technical-assessment.git
cd technical-assessment
```

HTTPS can be used instead if SSH is not configured:

```bash
git clone https://github.com/waltertaya/technical-assessment.git
```

### Configure the environment

Copy the example environment file and set the PostgreSQL connection string:

```bash
cp .env.example .env
```

Update `.env`:

```env
DATABASE_URL=postgres://user:password@localhost:5432/database_name?sslmode=disable
```

Make sure the PostgreSQL database in `DATABASE_URL` already exists. The migration creates the required tables and the `pgcrypto` extension:

```bash
go run ./internal/migrate
```

### Run the server

```bash
go run ./cmd/server
```

The API listens on `http://localhost:8080`. Verify that it is running:

```bash
curl http://localhost:8080/api/v1/health
```

Expected response:

```json
{"status":"healthy"}
```

### Run with Docker

Build the image from the project root:

```bash
docker build -t doctor-appointment .
```

Run the database migration before starting the application:

```bash
go run ./internal/migrate
```

Start the container with the environment variables from `.env`:

```bash
docker run --rm --name doctor-appointment \
	--env-file .env \
	-p 8080:8080 \
	doctor-appointment
```

The API is available at `http://localhost:8080`. When PostgreSQL runs outside the container, make sure `DATABASE_URL` uses a host reachable from Docker rather than `localhost`.

---

## API Routes

All routes are prefixed with `/api/v1` and accept or return JSON.

| Method | Route | Description |
| --- | --- | --- |
| `GET` | `/health` | Check API health |
| `POST` | `/doctors` | Add a doctor |
| `POST` | `/doctors/:id/working-hours` | Add a doctor's shift |
| `POST` | `/patients` | Add a patient |
| `POST` | `/appointments` | Book a 30-minute appointment |
| `GET` | `/doctors/:id/availability?date=YYYY-MM-DD` | Get available appointment slots |
| `PATCH` | `/appointments/:id/cancel` | Cancel an appointment |
| `PATCH` | `/appointments/:id/reschedule` | Reschedule an appointment |
| `GET` | `/patients/:id/appointments` | List a patient's upcoming appointments, sorted by date *(bonus)* |

### Validation rules

- A booking must fall within the doctor's working hours for that day.
- A booking cannot be made in the past, or within 1 hour of the current time *(bonus)*.
- A booking cannot overlap an existing, non-cancelled appointment.
- Cancelling an already-cancelled appointment returns an error.
- Rescheduling a cancelled appointment returns an error; the new slot is validated exactly as a fresh booking would be.

### Test with REST Client

Install the VS Code [REST Client extension](https://marketplace.visualstudio.com/items?itemName=humao.rest-client), start the API, and open [`requests.http`](requests.http) — it walks through creating a doctor, adding shifts (including a split morning/afternoon shift to exercise the midday-break logic), checking availability, booking, rescheduling, and cancelling.

---

## Run Tests

Run all Go tests from the project root:

```bash
go test ./...
```

For verbose output, coverage, race detection, or static analysis:

```bash
go test -v ./...
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
go test -race ./...
go vet ./...
```

Booking-logic tests exercise concurrent booking attempts directly to confirm the `FOR UPDATE` lock prevents double-booking under `-race`.

---

## Deployment & CI/CD

The repository includes [`.github/workflows/deploy.yml`](.github/workflows/deploy.yml), which deploys every push to `main` to an Ubuntu VPS over SSH. The workflow runs tests and `go vet`, builds the image, uploads it to the VPS, runs the migration, starts the container, and checks the health endpoint.

### VPS preparation

Install Docker and create the application directory on the VPS:

```bash
sudo apt update
sudo apt install -y docker.io curl
sudo systemctl enable --now docker
sudo mkdir -p /opt/doctor-appointment
sudo chown -R "$USER":"$USER" /opt/doctor-appointment
```

Create `/opt/doctor-appointment/.env` on the VPS. Do not commit this file:

```env
DATABASE_URL=postgres://user:password@database-host:5432/database_name?sslmode=disable
```

The VPS firewall/security group must allow SSH on the configured port and API traffic on port `8080` (or proxy the API through a web server). The deployment user must be able to run Docker without `sudo`:

```bash
sudo usermod -aG docker "$USER"
```

Log in again after running that command. The SSH user also needs write access to `/opt/doctor-appointment`.

### GitHub Actions secrets

Add these repository or `production` environment secrets in GitHub:

| Secret | Value |
| --- | --- |
| `VPS_HOST` | VPS public IP or DNS name, for example `my-vps-public-ip` |
| `VPS_PORT` | SSH port, normally `22` |
| `VPS_USER` | SSH user, for example `ubuntu` |
| `VPS_SSH_KEY` | Complete private key used to connect to the VPS |
| `VPS_KNOWN_HOSTS` | Output of `ssh-keyscan -p my-vps-public-ip` collected from a trusted machine |
| `VPS_APP_DIR` | Optional application directory; defaults to `/opt/doctor-appointment` |

The workflow uses the VPS-side `.env` for `DATABASE_URL`, so the database password is never written into the workflow or image artifact.

---

## AI Reflection

### 1. What did you use AI for across the four sections?

AI was used as a **brainstorming partner, autocomplete engine, and test-writing assistant** — not as the decision-maker on architecture. I drove every design choice; AI's role was to speed up execution once I'd already decided the approach:

- **Section 1 (System Design):** I used AI to sanity-check and brainstorm around a design I was already leaning toward — modeling working hours as discrete per-weekday shift rows instead of a single start/end column pair. I evaluated the alternatives myself and made the final call.
- **Section 2 (API Implementation):** AI's main contribution here was **autocompleting boilerplate** — repetitive Go struct definitions, handler scaffolding, and SQL migration syntax I'd already outlined — and **writing the initial test cases** for the booking logic, which I then reviewed, corrected, and extended (particularly the concurrency tests). The concurrency-safe transaction logic itself I designed and directed explicitly, after rejecting AI's first, naive, non-locking draft.
- **Section 3 (Deployment & CI/CD):** AI autocompleted the repetitive parts of the Dockerfile and GitHub Actions YAML syntax; I decided the pipeline stages, the deploy trigger, and the branch strategy.
- **Section 4 (AI Reflection):** I drafted this reflection myself; AI helped tighten the structure and wording.

### 2. Give one example where an AI suggestion improved your work. What did you prompt it with?

**The suggestion:** Modeling `working_hours` as discrete shift rows per weekday (rather than a single start/end pair per doctor) to handle midday breaks and irregular schedules cleanly.

**The prompt:**
> "I am building a clinic booking system in Go and PostgreSQL. The clinic has 5 doctors. We need to calculate 30-minute available slots dynamically. How should I model the working hours table to support flexible schedules, and what is the best strategy to handle midday lunch breaks without bloating the database?"

**The impact:** This validated the direction I was already exploring and helped me settle on a `working_hours` table with `day_of_week`, `start_time`, and `end_time`, allowing multiple shift rows per day per doctor (e.g., Monday 08:00–12:00 and Monday 13:00–17:00). I then designed the slot-generation algorithm myself to slice valid 30-minute intervals per shift row — giving zero database bloat and full scheduling flexibility, with midday breaks falling out naturally rather than needing special-case logic.

### 3. Give one example where AI output was wrong or incomplete and how you caught it.

**The failure:** AI's first draft of the `POST /appointments` handler used a plain, non-locking transaction: `SELECT` existing appointments, check availability in Go, then `INSERT`.

**How I caught it:** I recognized immediately that under concurrent load, two requests for the same slot could both run the `SELECT` before either `INSERT`, both see the slot as free, and both proceed — a textbook double-booking race condition. This wasn't something I needed AI to point out; it's a standard concurrency failure mode I know to check for in any booking/reservation system.

**The correction:** I rejected the naive flow and directed the implementation toward pessimistic locking: a `SELECT ... FOR UPDATE` row lock on the doctor's record inside the transaction boundary, forcing concurrent requests for the same doctor to serialize at the database level. I then wrote (and had AI help scaffold) concurrency tests specifically to prove the race was closed under `go test -race`.

### 4. Name two decisions you made without AI. Why did you trust your own judgment there?

**Decision A — Enforcing constraints at the database level, not just the application layer.**
I added native PostgreSQL constraints (`chk_times: start_time < end_time`, `0 ≤ day_of_week ≤ 6`, uniqueness/exclusion constraints on bookings) independent of any AI suggestion. AI-generated code tends to lean almost entirely on application-level validation. I trusted my own judgment here because defense-in-depth is a principle I hold regardless of tooling: if a future developer or another service bypasses the application layer, the database itself should still refuse to store invalid or conflicting data.

**Decision B — Using transactional rollbacks for test isolation instead of table truncation.**
I structured the integration test suite so each test runs inside its own SQL transaction that's rolled back on completion (`defer tx.Rollback()`), rather than truncating tables between tests — a pattern AI-generated test scaffolding tends to default to. I made this call myself because I know truncation is slow and blocks parallel test execution, while rollback-based isolation keeps the suite fast, side-effect-free, and safe to run concurrently in CI.

---

**Bottom line:** AI accelerated typing — boilerplate, repetitive syntax, first-pass test scaffolding — and served as a sounding board for ideas I was already forming. Every architectural and correctness-critical decision (schema design, locking strategy, constraint placement, test isolation strategy) was mine, and I caught and corrected the one place AI's first draft would have shipped a real bug.

---

## Author

[waltertaya](https://waltertaya.pages.dev/)

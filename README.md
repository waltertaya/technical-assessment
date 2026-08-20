# Doctor's Appointment
- Savannah Informatics Backend Developer Take-Home Assessment
- The system is implemented in *Go* using the *Gin Gonic* web framework and *PostgreSQL* as the persistent datastore.
- The architecture is designed to be highly maintainable, prevent booking conflicts (race conditions), and support real-world clinic constraints such as irregular shift structures and midday breaks.

## System Design

### Entity Relationship Diagram (ERD)
- To represent a highly realistic clinic setup without over-complicating storage requirements, we distinguish between standard working hour frames (shifts) and actual bookings. (instead of storing slots, we dynamically generate on-the-fly)
- *Working Hours (Shift)*: Represents active, non-overlapping daily working shifts for a doctor on specific days of the week. By allowing multiple rows per day of the week, this entity seamlessly accomodates midday breaks.
- *Appointment*: Represents a 30-minute booking reservation linking a `Patient` and a `Doctor` on a specific calendar date.

[![alt text](</src/images/savannah informatics OA - doctor appointment.png>)](https://dbdiagram.io/d/savannah-informatics-OA-doctor-appointment-6a85bacd6440800f52a25235)

### Dynamic Availability Calculation Logic
- Instead of pre-generating empty time slots in the database (which creates data-synchronization overhead and bloat), the system *calculates slot availability dynamically on-the-fly*.

[![alt text](/src/images/image.png)](https://lucid.app/lucidchart/1be14c71-1c7a-4cf4-992c-610fe8ce0024/edit?viewport_loc=110%2C-2343%2C1950%2C2396%2C0_0&invitationId=inv_d00e573a-96e4-4b5c-9b1a-99b1afebc5e3)

### Concurrency Handling & Race Conditions
- In high-traffic systems (scalability), multiple users might attempt to book the exact same 30-minute slot simultenously. If untreated, this leads to a **double-booking bug**
- The *Solution*: `Serializable Transactions & Pessimistic Locking`

### Architecture (above) Trade-offs
| Design Dimension | Selected Approach | Alternatives Considered | Trade-off Analysis |
| :------- | :------: | :-------: | ------- |
| Slot Generation | Dynamic Calculation | Static Pre-generation | `Pros`: Dynamic logic allows immediate adjustment of doctor working hours without retroactively editing rows. Decreases DB footprint.<br>`Cons`: Marginally higher CPU load on query endpoints (negated by Redis caching if needed in future expansion). |
| Concurrency Control | Database Lock (FOR UPDATE) | Go-Mutex / Distributed Locks | `Pros`: Robust and simple. Ensures consistency across multiple instances of the backend container without needing Redis Redlock orchestration.<br>`Cons`: Holds open transactional resources. For scale, locks are restricted exclusively to individual doctor scopes. |
| Data Integrity | DB-level Constraints | Application-level validation only | `Pros`: Guarantees absolute data integrity even if another microservice accesses the database or manual maintenance queries run.<br>`Cons`: Tight coupling between system rules and database engine schema. |

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

### Run tests

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

### Add a doctor

```bash
curl -X POST http://localhost:8080/api/v1/doctors \
	-H 'Content-Type: application/json' \
	-d '{"name":"Dr. Jane Doe","specialty":"Cardiology"}'
```

### Add a patient

```bash
curl -X POST http://localhost:8080/api/v1/patients \
	-H 'Content-Type: application/json' \
	-d '{"name":"John Smith","email":"john.smith@example.com"}'
```

### Add a doctor's working hours

`day_of_week` uses `0` for Sunday through `6` for Saturday. Multiple shifts can be added for the same day, which supports breaks during the day. Times use `HH:MM:SS`.

```bash
curl -X POST http://localhost:8080/api/v1/doctors/DOCTOR_ID/working-hours \
	-H 'Content-Type: application/json' \
	-d '{"day_of_week":1,"start_time":"09:00:00","end_time":"17:00:00"}'
```

### Book an appointment

Appointments are 30 minutes long. The date must be in the future and the requested start time must be in `HH:MM` format.

```bash
curl -X POST http://localhost:8080/api/v1/appointments \
	-H 'Content-Type: application/json' \
	-d '{"doctor_id":"DOCTOR_ID","patient_id":"PATIENT_ID","appointment_date":"2026-09-01","start_time":"09:00"}'
```

### Check availability

```bash
curl 'http://localhost:8080/api/v1/doctors/DOCTOR_ID/availability?date=2026-09-01'
```

### Cancel an appointment

```bash
curl -X PATCH http://localhost:8080/api/v1/appointments/APPOINTMENT_ID/cancel \
	-H 'Content-Type: application/json' \
	-d '{"reason":"Patient requested cancellation"}'
```

### Reschedule an appointment

```bash
curl -X PATCH http://localhost:8080/api/v1/appointments/APPOINTMENT_ID/reschedule \
	-H 'Content-Type: application/json' \
	-d '{"new_date":"2026-09-02","new_start_time":"10:30"}'
```

## Author
[waltertaya](https://waltertaya.pages.dev/)

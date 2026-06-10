# School Schedule API

REST API for a school schedule system. It supports student schedule views, teacher teaching schedules with total JP, admin JP recap reports, Excel import/export, API key authentication, and Swagger documentation.

Repository: <https://github.com/fadlnsyah/school-schedule-api>

## Tech Stack

- Go
- Gin
- GORM
- PostgreSQL, production target Supabase PostgreSQL
- Excelize for Excel import/export
- godotenv for local environment variables
- Google UUID
- Swaggo Swagger UI
- Render deployment

## Main Features

- API key authentication via `x-api-key`
- Schedule CRUD
- Student schedule by class and date
- Teacher schedule by date range with `total_jp`
- Admin JP recap by teacher
- Excel `.xlsx` upload with row-level validation
- Excel export for JP recap
- Class and teacher conflict detection
- Swagger UI at `/swagger/index.html`
- GORM AutoMigrate for `schedules`

## Folder Structure

```text
cmd/
config/
controllers/
docs/
dto/
middleware/
models/
routes/
services/
utils/
uploads/
reports/
tmp/
examples/
tools/
```

## Environment Variables

Copy `.env.example` to `.env` for local development.

```env
API_KEY=SECRET123
DATABASE_URL=postgresql://user:password@host:5432/database?sslmode=require
APP_PORT=8080
APP_ENV=development
```

Do not commit `.env`.

## Local Setup

```bash
go mod tidy
go run cmd/main.go
```

The app uses `PORT` first, then `APP_PORT`, then `8080`. Server binds to `0.0.0.0` through Gin's `:port` listener.

## Swagger

Generate Swagger docs:

```bash
swag init -g cmd/main.go
```

Open:

```text
http://localhost:8080/swagger/index.html
```

## Authentication

All `/api` endpoints require:

```http
x-api-key: SECRET123
```

Unauthorized response:

```json
{
  "error": "Unauthorized"
}
```

## Example Create Schedule

```bash
curl -X POST http://localhost:8080/api/schedules \
  -H "x-api-key: SECRET123" \
  -H "Content-Type: application/json" \
  -d '{
    "class_code": "XA01",
    "class_name": "X-A",
    "subject_code": "CHEM",
    "teacher_nik": "20222029",
    "teacher_name": "Najdin Aqmarina, S.Pd.",
    "date": "2025-02-10",
    "jam_ke": 2,
    "time_start": "08:40:00",
    "time_end": "09:20:00"
  }'
```

## Excel Upload

Endpoint:

```http
POST /api/schedules/upload
Content-Type: multipart/form-data
```

Form field: `file`

Required headers in the first Excel row:

```text
class_code, class_name, subject_code, teacher_nik, teacher_name, date, jam_ke, time_start, time_end
```

Sample file:

```text
examples/sample-schedules.xlsx
```

## Excel Export

```bash
curl -L "http://localhost:8080/api/schedules/export?start_date=2025-02-01&end_date=2025-02-28" \
  -H "x-api-key: SECRET123" \
  -o rekap-jp-2025-02-01-2025-02-28.xlsx
```

## Endpoint List

- `GET /health`
- `POST /api/schedules`
- `GET /api/schedules`
- `GET /api/schedules/{id}`
- `PUT /api/schedules/{id}`
- `DELETE /api/schedules/{id}`
- `GET /api/schedules/student?class_code=XA01&date=2025-02-10`
- `GET /api/schedules/teacher?teacher_nik=20222029&start_date=2025-02-10&end_date=2025-02-14`
- `GET /api/schedules/report/rekap-jp?start_date=2025-02-01&end_date=2025-02-28`
- `POST /api/schedules/upload`
- `GET /api/schedules/export?start_date=2025-02-01&end_date=2025-02-28`
- `GET /swagger/index.html`

## Conflict Detection

The API prevents schedule overlaps for the same class or teacher on the same date.

Overlap rule:

```text
new_time_start < existing_time_end
AND
new_time_end > existing_time_start
```

Create, update, and Excel upload use the same conflict check. On conflict, create/update returns `409 Conflict`; upload records the failed row in the `errors` array and continues inserting valid rows.

## Render Deployment

Build command:

```bash
go build -o main ./cmd/main.go
```

Start command:

```bash
./main
```

Set environment variables in Render:

- `API_KEY`
- `DATABASE_URL`
- `APP_ENV=production`

Render provides `PORT`; the app reads it automatically.

Live Swagger URL after deploy:

```text
https://your-render-service.onrender.com/swagger/index.html
```

## API Test File

Use `api-test.http` with VS Code REST Client for health check, CRUD, student, teacher, recap, upload, and export requests.

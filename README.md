# School Schedule API

REST API Sistem Jadwal Pelajaran Sekolah. API ini mendukung pengelolaan jadwal, tampilan jadwal siswa, jadwal mengajar guru, rekap Jam Pelajaran (JP) untuk Yayasan/Admin, import Excel, export Excel, dan dokumentasi Swagger.

Repository: <https://github.com/fadlnsyah/school-schedule-api>

## Live Deployment

- Base URL: <https://school-schedule-api-p7pg.onrender.com>
- Swagger UI: <https://school-schedule-api-p7pg.onrender.com/swagger/index.html>

## Tech Stack

- Go
- Gin
- GORM
- PostgreSQL
- Excelize
- Swagger
- Render

## Authentication

All `/api` endpoints require an API key header:

```http
x-api-key: SECRET123
```

Invalid or missing API key response:

```json
{
  "error": "Unauthorized"
}
```

## Main Features

- Schedule CRUD
- Student schedule by class and date
- Teacher schedule by date range with `total_jp`
- Foundation/Admin JP recap with per-class detail
- Excel `.xlsx` upload with row-level validation
- Excel export for JP recap
- Class and teacher schedule conflict detection
- Swagger UI at `/swagger/index.html`
- GORM AutoMigrate for the `schedules` table

## Endpoint List

Health:

- `GET /health`

CRUD Jadwal:

- `GET /api/schedules`
- `GET /api/schedules/{id}`
- `POST /api/schedules`
- `PUT /api/schedules/{id}`
- `DELETE /api/schedules/{id}`

Upload & Export Excel:

- `POST /api/schedules/upload`
- `GET /api/schedules/export?start_date=2025-02-01&end_date=2025-02-28`

Frontend Siswa:

- `GET /api/schedules/student?class_code=XA01&date=2025-02-10`

Frontend Guru:

- `GET /api/schedules/teacher?teacher_nik=20222029&start_date=2025-02-10&end_date=2025-02-14`

Frontend Yayasan/Admin:

- `GET /api/schedules/report/rekap-jp?start_date=2025-02-01&end_date=2025-02-28`

Swagger:

- `GET /swagger/index.html`

## Example Curl Commands

Get all schedules:

```bash
curl -X GET "https://school-schedule-api-p7pg.onrender.com/api/schedules" \
  -H "x-api-key: SECRET123"
```

Create schedule:

```bash
curl -X POST "https://school-schedule-api-p7pg.onrender.com/api/schedules" \
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

Get student schedule:

```bash
curl -X GET "https://school-schedule-api-p7pg.onrender.com/api/schedules/student?class_code=XA01&date=2025-02-10" \
  -H "x-api-key: SECRET123"
```

Get teacher schedule:

```bash
curl -X GET "https://school-schedule-api-p7pg.onrender.com/api/schedules/teacher?teacher_nik=20222029&start_date=2025-02-10&end_date=2025-02-14" \
  -H "x-api-key: SECRET123"
```

Get Foundation/Admin JP recap:

```bash
curl -X GET "https://school-schedule-api-p7pg.onrender.com/api/schedules/report/rekap-jp?start_date=2025-02-01&end_date=2025-02-28" \
  -H "x-api-key: SECRET123"
```

Upload Excel:

```bash
curl -X POST "https://school-schedule-api-p7pg.onrender.com/api/schedules/upload" \
  -H "x-api-key: SECRET123" \
  -F "file=@schedules.xlsx"
```

Export Excel:

```bash
curl -X GET "https://school-schedule-api-p7pg.onrender.com/api/schedules/export?start_date=2025-02-01&end_date=2025-02-28" \
  -H "x-api-key: SECRET123" \
  --output rekap-jp.xlsx
```

The exported file contains this table format:

```text
No | NIK | Nama Pengajar | Kelas yg Diajar | Pekan 1 | Pekan 2 | Pekan 3 | Pekan 4 | Pekan 5 | Total JP
```

Weekly JP is calculated from the day of month:

- Pekan 1: day 1-7
- Pekan 2: day 8-14
- Pekan 3: day 15-21
- Pekan 4: day 22-28
- Pekan 5: day 29-31

## Foundation/Admin Recap Response

`GET /api/schedules/report/rekap-jp` returns:

```json
{
  "periode": {
    "start_date": "2025-02-01",
    "end_date": "2025-02-28"
  },
  "total_pengajar": 3,
  "rekap": [
    {
      "teacher_nik": "20222029",
      "teacher_name": "Najdin Aqmarina, S.Pd.",
      "total_jp": 40,
      "total_kelas": 3,
      "detail": [
        {
          "class_code": "XA01",
          "class_name": "X-A",
          "jumlah_jp": 15
        }
      ]
    }
  ]
}
```

If no schedules exist for the selected period, the API still returns `200 OK`:

```json
{
  "periode": {
    "start_date": "2025-02-01",
    "end_date": "2025-02-28"
  },
  "total_pengajar": 0,
  "rekap": []
}
```

## Environment Variables

Required environment variables:

- `DATABASE_URL`
- `API_KEY`
- `PORT`

Local development also supports:

- `APP_PORT`
- `APP_ENV`

Example:

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

The app reads `PORT` first, then `APP_PORT`, then defaults to `8080`.

## Swagger

Generate Swagger docs:

```bash
swag init -g cmd/main.go
```

Local Swagger:

```text
http://localhost:8080/swagger/index.html
```

Live Swagger:

```text
https://school-schedule-api-p7pg.onrender.com/swagger/index.html
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

## Conflict Detection

The API prevents schedule overlaps for the same class or teacher on the same date.

Overlap rule:

```text
new_time_start < existing_time_end
AND
new_time_end > existing_time_start
```

Create, update, and Excel upload use the same conflict check. On conflict, create/update returns `409 Conflict`; upload records the failed row in the `errors` array and continues inserting valid rows.

## Deployment Notes

This project is deployed to Render and uses an online PostgreSQL database. Render provides `PORT` automatically, and the app reads it before `APP_PORT`.

Render build command:

```bash
go build -o main ./cmd/main.go
```

Render start command:

```bash
./main
```

## Testing Notes

If `GET /api/schedules` returns:

```json
{
  "data": [],
  "message": "Schedules retrieved successfully",
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 0
  }
}
```

then the API is accessible, but the database is still empty.

Use `api-test.http` with VS Code REST Client for local API testing.

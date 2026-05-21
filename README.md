# HR Cloud Service

A small Go backend example for HR-style employee management. The project uses a traditional MVC-style structure, close to what you may know from Java/Spring:

- `cmd/api`: application entrypoint
- `internal/model`: data models and request DTOs
- `internal/controller`: HTTP controllers
- `internal/service`: business logic
- `internal/repository`: data access layer
- `internal/server`: route registration and server wiring

Request flow:

```text
HTTP request -> Controller -> Service -> Repository -> Model
```

## Run

Install Go, then run:

```powershell
cd D:\GoLang\hr-cloud-service
go run .\cmd\api
```

The API listens on `http://localhost:8080`.

## Endpoints

```http
GET  /healthz
GET  /api/v1/employees
POST /api/v1/employees
GET  /api/v1/employees/{id}
```

Example create request:

```json
{
  "name": "Nguyen Van A",
  "email": "a@example.com",
  "department": "Engineering",
  "title": "Backend Engineer"
}
```

## Note

This is MVC-style, but adapted to Go. The service layer keeps business rules out of controllers, and the repository layer keeps data access replaceable when you move from in-memory storage to PostgreSQL, MySQL, MongoDB, or Redis.

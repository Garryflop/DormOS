# Issue Service — Ernar

Microservice for managing maintenance requests (issues) in a dormitory. Part of the **DormOS** project.

## 📋 Overview

Issue Service handles the full lifecycle of resident requests:
- Creating and viewing issues
- Status management (open → in_progress → resolved)
- Assigning workers to issues
- Comments on issues
- Category management
- Event publishing via NATS on creation and status change

## 🏗 Stack

| Technology | Usage |
|---|---|
| Go 1.25 | Main language |
| gRPC + Protobuf | Inter-service communication |
| PostgreSQL 15 | Data storage |
| Redis | Caching |
| NATS | Event Bus |
| Docker | Containerization |

## 📁 Structure

```
issue-service/
├── cmd/
│   └── main.go                  # Entry point, initialization
├── internal/
│   ├── domain/
│   │   └── issue.go             # Data models
│   ├── repository/
│   │   ├── issue_repo.go        # CRUD for issues
│   │   └── other_repos.go       # Comments, workers, categories
│   ├── service/
│   │   └── issue_service.go     # Business logic (12 methods)
│   ├── grpc/
│   │   └── server.go            # gRPC server
│   └── nats/
│       └── publisher.go         # NATS publisher
├── migrations/
│   ├── create_issue.up.sql      # Create tables
│   └── create_issue.down.sql    # Rollback migrations
├── Dockerfile
└── go.mod
```

## 🗄 Database

### Tables

**`issue_categories`** — issue categories (Plumbing, Electrical, etc.)

**`workers`** — workers who get assigned to issues

**`issues`** — main issues table
| Field | Type | Description |
|---|---|---|
| id | UUID | Primary key |
| user_id | UUID | Resident |
| room_number | VARCHAR | Room number |
| category_id | UUID | Category |
| title | VARCHAR | Issue title |
| description | TEXT | Problem description |
| status | VARCHAR | open / in_progress / resolved / closed |
| worker_id | UUID | Assigned worker |
| photo_url | VARCHAR | Photo of the problem |

**`issue_comments`** — comments on issues

## 🔌 gRPC Endpoints (12 total)

| # | Method | Role | Description |
|---|---|---|---|
| 1 | `CreateIssue` | student+ | Create an issue |
| 2 | `GetIssue` | student+ | Get issue by ID |
| 3 | `ListMyIssues` | student | My issues |
| 4 | `ListAllIssues` | manager+ | All issues |
| 5 | `UpdateIssueStatus` | manager+ | Change status |
| 6 | `DeleteIssue` | admin | Delete an issue |
| 7 | `AddComment` | student+ | Add a comment |
| 8 | `ListComments` | student+ | List comments |
| 9 | `AssignWorker` | manager+ | Assign a worker |
| 10 | `ListWorkers` | manager+ | List available workers |
| 11 | `CreateCategory` | admin | Create a category |
| 12 | `ListCategories` | public | List all categories |

## 📨 NATS Events

| Event | When | Listener |
|---|---|---|
| `issue.created` | Issue created | Notification Service |
| `issue.status_changed` | Status updated | Notification Service |

## 🚀 Running

### Via Docker Compose (recommended)

```bash
# From DormOS root
docker-compose up -d --build issue-service
docker-compose logs -f issue-service
```

### Locally

```bash
cd issue-service
export DATABASE_URL="postgres://dormos:dormos123@localhost:5432/dormos?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export REDIS_PASSWORD="redis123"
export NATS_URL="nats://localhost:4222"
go run cmd/main.go
```

### Migrations

```bash
# PowerShell
Get-Content migrations/create_issue.up.sql | docker exec -i dormos-postgres-1 psql -U dormos -d dormos
```

## 🔧 Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | postgres://dormos:dormos123@... | PostgreSQL connection |
| `REDIS_ADDR` | localhost:6379 | Redis address |
| `REDIS_PASSWORD` | redis123 | Redis password |
| `NATS_URL` | nats://localhost:4222 | NATS URL |
| `GRPC_PORT` | 50052 | gRPC port |

## 🧪 Testing via Postman

1. Open Postman → New → gRPC Request
2. URL: `localhost:50052`
3. Import `issue.proto`
4. Select a method and send

Example `CreateIssue`:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "room_number": "305",
  "category_id": "<id from ListCategories>",
  "title": "Broken tap",
  "description": "Water is leaking constantly from the tap"
}
```

## 👤 Author

**Ernar** — Issue & Maintenance Service + React Frontend
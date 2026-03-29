# go-concurrency-sql-lab

High-Performance E-Commerce Analytics — **Backend:** Go 1.22+ · **Database:** PostgreSQL

**Purpose**: This repository is created as a hands-on lab to master **Go Routines** (basic to advanced patterns like Worker Pools) and **SQL Analytics** (specifically mastering `GROUP BY` and `HAVING` clauses) by simulating a high-concurrency e-commerce order ingestion and analytics pipeline.

## Documentation layout

| Layer | Purpose | Avoid |
|------|---------|--------|
| **[README.md](./README.md)** (this file) | Single entry to the repo: overview, quick start, folder map, links to `docs/` | Pasting long content from `docs/` or duplicating full API tables |
| **[docs/](./docs/)** | In-depth **canonical** docs — architecture, flows, planning, testing | — |
| **[execute.md](./execute.md)** | Phase / progress checklist | — |
| **[RULE.md](./RULE.md)** | Rules for updating `.md` when code changes affect structure or behavior described in docs | — |

This matches common monorepo practice: **repo home = navigation**, details live in `docs/`, subpackages have short READMEs.

## Quick start

### Environment Variables

From the **repository root**:

```bash
cp .env.example .env
```
*(You will need to create `.env.example` to define Database configurations)*

### Running the Database

Start the PostgreSQL database via Docker Compose:

```bash
docker-compose up -d db
```

### Running the Application

Download Go modules:

```bash
go mod tidy
```

Run the API / Background worker:

```bash
go run cmd/api/main.go
```

### Data Seeding

To run the concurrent worker pool for bulk inserting 1,000,000 fake orders:

```bash
go run cmd/seed/main.go
```

## Documentation (read more)

| Topic | Link |
|--------|------|
| Full documentation index | [docs/README.md](./docs/README.md) |
| Architecture & tech stack | [docs/architech.md](./docs/architech.md) |
| Future plans & roadmap | [docs/planning.md](./docs/planning.md) |
| Progress / phases | [execute.md](./execute.md) |
| Doc–code sync rules | [RULE.md](./RULE.md) |

## Repository layout

| Folder | Description |
|----------|----------|
| [`cmd/`](./cmd/) | Executable entrypoints (e.g., `cmd/api`, `cmd/seed`) |
| [`internal/`](./internal/) | Core business logic, SQL repositories, and Goroutine handlers |
| [`docs/`](./docs/) | Design docs, architecture, and planning |
| [`scripts/`](./scripts/) | SQL scripts for initialization or testing |
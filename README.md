# Project grpc-product

**This is a Go web application built with the [Gin](https://github.com/gin-gonic/gin) framework, Gorm, and PostgreSQL as the database. This project supports running either via Docker or locally with native Go tooling.**
## Getting Started

## Run Locally with Go Build
### 1. Prerequisites
  - Go (v1.22 or newer)

  - PostgreSQL

### 2. Start PostgreSQL Locally
If you're not using Docker, start a local Postgres instance manually (or use a tool like pgAdmin, Postgres.app, or local service manager). Make sure it's using the same credentials defined in the .env file.

### 3. Set Environment Variables
```env
DB_DATABASE=db_database
DB_USERNAME=postgres
DB_PASSWORD=secret
DB_PORT=5432
DB_HOST=localhost
PORT=8080
APP_ENV=local
```

### 4. Build and run app
```bash
go build -o main ./cmd/server
./main
```


## Run using MakeFile

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```
Create DB container
```bash
make docker-run
```

Shutdown DB Container
```bash
make docker-down
```

DB Integrations Test:
```bash
make itest
```

Live reload the application:
```bash
make watch
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```

## 🐳 Run with Docker Compose

### 1. Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/)

### 2. Setup Environment Variables

Create a `.env` file at the root of the project with the following content:

```env
DB_DATABASE=db_database
DB_USERNAME=postgres
DB_PASSWORD=secret
DB_PORT=5432
DB_HOST=localhost
PORT=8080
APP_ENV=local
```

### 3. Build and start the services:
```bash
docker-compose up --build
```

### 4. Stopping the services:
```bash
docker-compose down
```


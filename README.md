# RDEV Boilerplate

This repository provides a simplified boilerplate for a RESTful API built with [Golang](https://golang.org/), [Gin](https://github.com/gin-gonic/gin) (a web framework for Go), and [Bun](https://github.com/uptrace/bun) (a SQL ORM for PostgreSQL). This template is designed to get you up and running quickly with basic API functionality.

## Features

- Simple REST API structure
- Gin for HTTP routing and middleware
- Bun ORM for PostgreSQL integration
- Basic database model setup
- Example CRUD operations for a resource
- Environment-based configuration management using `yml` file

## Tech Stack

- **Go (Golang)**: Programming language used to build the API.
- **Gin**: Web framework for building fast and efficient web applications.
- **Bun**: PostgreSQL ORM for Go, enabling interaction with a PostgreSQL database.
- **PostgreSQL**: Relational database used for data persistence.

## Prerequisites

Before running the project, ensure you have the following installed:

- [Go (1.25+)](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)
- [Git](https://git-scm.com/)
- [Go Modules](https://blog.golang.org/using-go-modules)
- (Optional): [Docker](https://www.docker.com/)

## Setup

### 1. Clone the Repository

```bash
git clone https://github.com/XaiPhyr/rdev_boilerplate
cd rdev_boilerplate
```

### 2. Install Dependencies

```bash
cd api
go mod tidy
```

### 3. Create Database

```sql
CREATE DATABASE your_db_name;
```

### 4. Configure config.yml file
remove .example on conf/config.yml.example

```yml
env: 'dev'

server:
  host: 'localhost:8200'
  endpoint: '/v1/api'
  jwt_key: ''

database:
  dsn: 'postgres://postgres:postgres@localhost:5432/your_db?sslmode=disable&TimeZone=Asia/Singapore'
  bundebug: true
```

### 5. SQL tables
SQL files can be found on sql/

### 6. Run the API

```bash
go run .
```

By default, the API will be accessible at http://localhost:8200.
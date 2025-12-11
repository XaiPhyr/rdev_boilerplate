# RDEV Boilerplate

This repository offers a streamlined boilerplate for creating a RESTful API using Golang, Gin (a web framework for Go), and Bun (a SQL ORM for PostgreSQL). The goal is to provide a simple, minimal-code setup that allows you to quickly launch an API with essential functionality.

## Features

- Simple REST API structure
- Gin for HTTP routing and middleware
- Bun ORM for PostgreSQL integration
- Basic database model setup
- Example **UPSERT**, **READ**, and **DELETE** operations for a resource
- Environment-based configuration management using `yml` file
- **Role-Based Access Control (RBAC)** for user roles and permissions management
  - Includes **Group Access Control** for user grouping and role-specific permissions
- Rate limiting middleware to control request throughput per client

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
- (Optional, coming soon): [Docker](https://www.docker.com/)

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

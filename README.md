# Server Reservation System

A RESTful API service for managing server reservations.

## Features

- Server management (add, remove, list servers)
- Reservation management (create, cancel, list reservations)
- User authentication
- SQLite database for data persistence

## Prerequisites

- Go 1.21 or higher
- SQLite3

## Installation

1. Clone the repository:
```bash
git clone https://github.com/rusik69/serverscheduler.git
cd serverscheduler
```

2. Install dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go
├── database/
│   └── database.go
├── handlers/
│   └── handlers.go
├── middleware/
│   └── auth.go
├── models/
│   └── models.go
├── go.mod
└── README.md
```

## API Endpoints

### Authentication
- POST /api/auth/register - Register a new user
- POST /api/auth/login - Login and get JWT token

### Servers
- GET /api/servers - List all servers
- POST /api/servers - Add a new server
- DELETE /api/servers/:id - Remove a server

### Reservations
- GET /api/reservations - List all reservations
- POST /api/reservations - Create a new reservation
- DELETE /api/reservations/:id - Cancel a reservation

## Database Schema

The application uses SQLite with the following tables:
- users
- servers
- reservations 
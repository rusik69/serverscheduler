# Server Scheduler

A server reservation system built with Go backend and Vue.js frontend.

## Features

- User authentication and authorization
- Server management (admin only)
- Reservation scheduling
- Password change functionality for all users
- **Server credential display for root users** - Root users can view server passwords and credentials
- Visual calendar interface for server availability
- RESTful API
- Modern web interface

## Project Structure

```
serverscheduler/
├── cmd/server/          # Main application entry point
├── internal/            # Internal packages
│   ├── database/        # Database operations
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # Authentication middleware
│   └── models/          # Data models
├── frontend/            # Vue.js frontend application
│   ├── src/            # Vue.js source code
│   ├── public/         # Static assets
│   ├── Dockerfile      # Frontend Docker configuration
│   └── nginx.conf      # Nginx configuration
├── Dockerfile.backend  # Backend Docker configuration
├── docker-compose.yml  # Multi-service orchestration
└── Makefile            # Build automation
```

## Development

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker (optional)

### Backend Development

```bash
# Install dependencies
go mod download

# Run tests
make test

# Run development server
make dev-backend
# or
go run ./cmd/server
```

### Frontend Development

```bash
# Install dependencies and run development server
make dev-frontend
# or
cd frontend && npm install && npm run serve
```

## Docker Deployment

### Option 1: Docker Compose (Recommended)

Run both frontend and backend services together:

```bash
# Build and start all services
make docker-compose-up

# View logs
make docker-compose-logs

# Stop services
make docker-compose-down

# Rebuild services
make docker-compose-build
```

Services will be available at:
- Frontend: http://localhost
- Backend API: http://localhost:8080

### Option 2: Individual Docker Builds

Build and run services separately for more control:

```bash
# Build backend image
make docker-build-backend

# Build frontend image
make docker-build-frontend

# Build both images
make docker-build-all

# Run backend
docker run -d -p 8080:8080 -v $(PWD)/data:/app/data ghcr.io/rusik69/serverscheduler-backend:latest

# Run frontend
docker run -d -p 80:80 ghcr.io/rusik69/serverscheduler-frontend:latest
```

## Configuration

### Environment Variables

- `PORT` - Server port (default: 8080)
- `DB_PATH` - SQLite database path (default: data/serverscheduler.db)
- `ROOT_PASSWORD` - Set a specific root password (optional)
- `RESET_ROOT_PASSWORD` - Set to `true` to reset root password on startup (optional)

### Root User Management

On first startup, a root user is automatically created:
- Username: `root`
- Password: (randomly generated and displayed in logs)

#### Managing Root Password

**If root user already exists:**
- The application will show if `ROOT_PASSWORD` is set in environment variables
- To reset the password, set `RESET_ROOT_PASSWORD=true` and restart
- To use a specific password, set `ROOT_PASSWORD=your_password` and reset

**Examples:**
```bash
# Set specific root password
export ROOT_PASSWORD="mySecurePassword123"

# Reset root password with random generation
export RESET_ROOT_PASSWORD=true

# Reset root password with specific password
export ROOT_PASSWORD="newPassword456"
export RESET_ROOT_PASSWORD=true
```

#### Password Change via Frontend

All authenticated users (including root) can change their password through the web interface:
1. Click on the user dropdown in the top navigation
2. Select "Change Password"
3. Enter current password and new password
4. New password must be at least 6 characters and different from current password

#### Server Credential Access (Root Users Only)

Root users have access to server credentials for administrative purposes:

**In Server Management:**
- Server passwords are displayed in a dedicated column
- Click on password tags to copy credentials to clipboard
- Toggle password visibility with show/hide button
- Credentials are hidden from non-root users

**In Reservations:**
- Root users see server access information for all reservations
- Includes server IP address, username, and password
- Facilitates server access during reservation periods
- Regular users only see reservation details without credentials

**Security Notes:**
- Server passwords are stored in plaintext for administrative access
- Only root users can view server credentials
- Password fields include copy-to-clipboard functionality
- All credential access is logged for audit purposes

## Architecture

### Backend (Go)
- **Gin** - HTTP web framework
- **SQLite** - Database
- **JWT** - Authentication
- **bcrypt** - Password hashing

### Frontend (Vue.js)
- **Vue 3** - Progressive framework
- **Element Plus** - UI components
- **Vuex** - State management
- **Vue Router** - Client-side routing
- **Axios** - HTTP client

### Deployment
- **Docker** - Containerization
- **Nginx** - Frontend web server and reverse proxy
- **Docker Compose** - Multi-service orchestration

## Development Commands

```bash
# Backend
make build          # Build backend binary
make run            # Run backend
make test           # Run backend tests
make dev-backend    # Run backend in development mode

# Frontend
make frontend-build # Build frontend for production
make frontend-serve # Run frontend development server
make dev-frontend   # Run frontend in development mode
make frontend-test  # Run frontend tests
make frontend-test-watch    # Run frontend tests in watch mode
make frontend-test-coverage # Generate frontend test coverage

# Testing
make test-all       # Run both backend and frontend tests

# Docker
make docker-build-backend    # Build backend Docker image
make docker-build-frontend   # Build frontend Docker image
make docker-build-all        # Build both images
make docker-push-backend     # Build and push backend image
make docker-push-frontend    # Build and push frontend image
make docker-push-all         # Build and push both images
make docker-compose-up       # Start all services
make docker-compose-down     # Stop all services
make docker-compose-logs     # View service logs
make docker-clean           # Clean up Docker resources
```

## Testing

The project includes comprehensive tests for both backend and frontend:

### Backend Tests
- Database operations and migrations
- HTTP handlers and API endpoints
- Authentication middleware and JWT handling
- User management and authorization
- Server and reservation CRUD operations
- Password change functionality

### Frontend Tests
- Vue.js component testing with Vue Test Utils
- Vuex store modules (authentication, state management)
- API client configuration and HTTP interceptors
- View components (Login, Servers, Reservations, Calendar)
- Form validation and user interactions
- Error handling and edge cases
- Role-based access control

### Test Framework
- **Backend**: Go testing framework with testify
- **Frontend**: Jest with Vue Test Utils and jsdom
- **API Mocking**: axios-mock-adapter for HTTP request mocking
- **Coverage**: Automated coverage reporting via CI/CD

### Running Tests

```bash
# Run all tests (backend + frontend)
make test-all

# Run only backend tests
make test

# Run only frontend tests
make frontend-test

# Run frontend tests in watch mode
make frontend-test-watch

# Generate frontend test coverage report
make frontend-test-coverage
```

### Test Coverage Goals
- **Backend**: > 80% coverage for handlers, models, and core logic
- **Frontend**: > 80% statements, > 75% branches

### CI/CD Integration
Tests run automatically on:
- Every push to main/develop branches
- Every pull request
- Coverage reports uploaded to Codecov
- All tests must pass before deployment

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login
- `GET /api/auth/user` - Get current user info
- `POST /api/auth/change-password` - Change user password (authenticated users)

### Servers (Protected)
- `GET /api/servers` - List all servers
- `POST /api/servers` - Create server (admin only)
- `GET /api/servers/:id` - Get server details
- `PUT /api/servers/:id` - Update server
- `DELETE /api/servers/:id` - Delete server

### Reservations (Protected)
- `GET /api/reservations` - List user's reservations
- `POST /api/reservations` - Create reservation
- `GET /api/reservations/:id` - Get reservation details

### Password Management
- `POST /api/auth/change-password` - Change current user's password

**Request body for password change:**
```json
{
  "current_password": "currentPassword123",
  "new_password": "newSecurePassword456"
}
```

**Requirements:**
- User must be authenticated
- Current password must be correct
- New password must be at least 6 characters
- New password must be different from current password
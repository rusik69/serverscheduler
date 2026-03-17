# Server Scheduler

Web app for scheduling server access. Users reserve time slots, admins manage servers and users. Supports SSH key auth and Slack notifications.

## Setup

```bash
cp .env.example .env
# Edit .env - set ADMIN_PASSWORD and optionally SLACK_WEBHOOK_URL
```

## Run

**Local:**
```bash
make dev
```

**Podman (local testing):**
```bash
make podman-run
```

**Docker Compose:**
```bash
make docker-compose-up
```

App runs at http://localhost:8080

## Make targets

| Target | Description |
|--------|-------------|
| `make dev` | Build and run locally |
| `make test` | Run tests |
| `make podman-run` | Build and run with Podman |
| `make podman-stop` | Stop Podman container |
| `make docker-compose-up` | Start with Docker Compose |
| `make docker-compose-down` | Stop Docker Compose |
| `make deploy-deps` | Install Docker on remote host (run before first deploy) |
| `make deploy` | Deploy to host via SSH (`DEPLOY_HOST=user@host`, `DEPLOY_PATH` defaults to `~/serverscheduler`; requires Docker on remote) |

## Configuration

| Variable | Description |
|----------|-------------|
| `PORT` | HTTP port (default: 8080) |
| `DB_PATH` | SQLite path |
| `ADMIN_USERNAME` | Admin login |
| `ADMIN_PASSWORD` | Admin password (required) |
| `SLACK_WEBHOOK_URL` | Optional Slack notifications |
| `LOG_LEVEL` | Log level (default: info) |

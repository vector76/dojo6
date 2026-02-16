# Dojo CRM

A management application for martial arts and fitness studios. Handles member management, class scheduling, attendance tracking, and payment recording.

## Prerequisites

- Go 1.24+
- Node.js 18+
- npm

## Quick Start

```bash
# Build the single binary (compiles frontend + backend)
make build

# Run the server
./dojo6

# Open http://localhost:8080 in your browser
# On first visit, you'll be prompted to create an admin account
```

## Development

```bash
# Install frontend dependencies
make frontend-install

# Start dev servers (Go backend + Vite dev server with hot reload)
make dev
# Backend: http://localhost:8080
# Frontend dev: http://localhost:5173 (proxies /api to backend)

# Run all tests
make test

# Clean build artifacts
make clean
```

## Environment Variables

| Variable     | Default    | Description                                      |
|-------------|------------|--------------------------------------------------|
| `PORT`      | `8080`     | Server listen port                               |
| `DB_PATH`   | `dojo6.db` | SQLite database file path                        |
| `JWT_SECRET` | (random)  | JWT signing secret; random if not set (tokens won't survive restarts) |

## Deployment

The `make build` command produces a single self-contained binary (`dojo6`) with the React frontend embedded. To deploy:

1. Build: `make build`
2. Copy the `dojo6` binary to your server
3. Set `JWT_SECRET` to a stable secret value
4. Run: `JWT_SECRET=your-secret ./dojo6`

The SQLite database file is created automatically on first run.

## Architecture

- **Backend**: Go with chi router, SQLite (modernc.org/sqlite), JWT auth
- **Frontend**: React 19 + TypeScript + Vite, embedded in the Go binary via `//go:embed`
- **Auth**: JWT tokens with role-based access control (admin, instructor, user)

## User Roles

| Role       | Capabilities                                                        |
|------------|---------------------------------------------------------------------|
| Admin      | Full access: manage members, classes, class types, attendance, payments |
| Instructor | View members, manage classes and attendance                          |
| User       | View own attendance and payment history                              |

## API Endpoints

### Auth
- `POST /api/auth/login` - Login
- `POST /api/auth/setup` - Initial admin setup
- `GET /api/auth/setup-status` - Check if setup is needed
- `GET /api/auth/me` - Current user info (authenticated)

### Members (admin/instructor)
- `GET /api/users` - List members
- `POST /api/users` - Create member (admin)
- `GET /api/users/:id` - Get member
- `PUT /api/users/:id` - Update member
- `DELETE /api/users/:id` - Soft-delete member (admin)

### Classes
- `GET /api/classes` - List classes (authenticated)
- `POST /api/classes` - Create class (admin)
- `PUT /api/classes/:id` - Update class (admin)
- `DELETE /api/classes/:id` - Delete class (admin)

### Class Types
- `GET /api/class-types` - List class types (authenticated)
- `POST /api/class-types` - Create class type (admin)
- `PUT /api/class-types/:id` - Update class type (admin)
- `DELETE /api/class-types/:id` - Delete class type (admin)

### Attendance
- `POST /api/classes/:id/attendance` - Record attendance (admin/instructor)
- `GET /api/classes/:id/attendance` - List class attendance (admin/instructor)
- `GET /api/users/:id/attendance` - User attendance history (self/admin/instructor)

### Payments
- `POST /api/payments` - Record payment (admin)
- `GET /api/users/:id/payments` - User payment history
- `GET /api/users/:id/balance` - User balance info

# Ticket System (Golang Backend Intern Assignment)

A small backend service where a user can register, log in, create tickets,
view only their own tickets, and update the status of their own tickets.

## Tech / Design Notes

- **Language:** Go 1.22. Routing uses the standard library's `net/http`
  with Go 1.22's built-in method+path pattern matching (e.g.
  `GET /tickets/{id}`) — no third-party router/framework needed for
  something this size.
- **Storage:** In-memory, guarded by a `sync.RWMutex` (explicitly allowed
  by the assignment scope). Data resets when the process restarts.
- **Auth:** JWT (HS256) via [`golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt),
  the most widely used Go JWT library. The verifier explicitly pins the
  expected signing method (`*jwt.SigningMethodHMAC`) to guard against
  algorithm-confusion attacks, and validates both `exp` and that the
  subject/username claims are present.
- **Password hashing:** Passwords are never stored in plaintext — they're
  hashed with `golang.org/x/crypto/bcrypt` (cost factor 12), the standard,
  battle-tested choice for password hashing in Go. bcrypt generates and
  embeds its own per-password salt automatically.

## Project Structure

```
main.go        - server setup, config, routing
models.go      - User/Ticket structs, status transition rules
store.go       - in-memory data store
auth.go        - bcrypt password hashing + golang-jwt token issuing/verification
middleware.go  - Bearer-token auth middleware
auth_handlers.go - http handlers for all auth endpoints
ticket_handlers.go - http handlers for all ticket endpoints
go.mod/go.sum  - dependencies (golang-jwt/jwt, golang.org/x/crypto/bcrypt)
Dockerfile     - multi-stage build
.env.example   - example environment variables
```

## API Endpoints

| Method | Endpoint                | Auth required | Purpose                        |
|--------|--------------------------|---------------|---------------------------------|
| GET    | `/health`                | No            | Health check                   |
| POST   | `/auth/register`         | No            | Register a user                |
| POST   | `/auth/login`            | No            | Log in, returns a JWT           |
| POST   | `/tickets`               | Yes           | Create a ticket                 |
| GET    | `/tickets`               | Yes           | List the logged-in user's tickets |
| GET    | `/tickets/{id}`          | Yes           | Get one of your own tickets by ID |
| PATCH  | `/tickets/{id}/status`   | Yes           | Update the status of your own ticket |

Protected endpoints require an `Authorization: Bearer <token>` header.

### Request / response examples

**Register**
```
POST /auth/register
{"username": "alice", "password": "secret123"}
-> 201 {"id": "...", "username": "alice"}
```

**Login**
```
POST /auth/login
{"username": "alice", "password": "secret123"}
-> 200 {"token": "eyJhbGciOi..."}
```

**Create ticket**
```
POST /tickets
Authorization: Bearer <token>
{"title": "Printer broken", "description": "office printer jam"}
-> 201 {"id": "...", "user_id": "...", "title": "...", "description": "...", "status": "open", "created_at": "...", "updated_at": "..."}
```

**Update status**
```
PATCH /tickets/{id}/status
Authorization: Bearer <token>
{"status": "in_progress"}
-> 200 {... "status": "in_progress" ...}
```

### Status flow

- `open` -> `in_progress` or `closed`
- `in_progress` -> `closed`
- `closed` is terminal: it can never move back to `open` or `in_progress`,
  and any transition to the same status is rejected (no no-op updates).
- Invalid transitions return `400 Bad Request`.

### Ownership & error codes

- Requests without a valid `Authorization: Bearer <token>` header get `401`.
- Fetching or updating a ticket that exists but belongs to another user
  returns `403 Forbidden`.
- Fetching or updating a ticket ID that doesn't exist at all returns
  `404 Not Found`.
- Malformed request bodies / missing required fields return `400`.
- Registering an already-taken username returns `409 Conflict`.

## Local Run (without Docker)

Requires Go 1.22+.

```bash
go build -o ticket-system .
JWT_SECRET=some-long-random-secret ./ticket-system
curl http://localhost:8080/health
```

## Local Run (Docker) — matches the assignment's contract

```bash
docker build -t ticket-system .
docker run -p 8080:8080 -e JWT_SECRET=some-long-random-secret ticket-system
curl http://localhost:8080/health
```

Expected health response:
```json
{"status": "ok"}
```

## Environment Variables

See `.env.example`.

| Variable     | Required | Default                  | Description                       |
|--------------|----------|---------------------------|------------------------------------|
| `JWT_SECRET` | Recommended | `dev-secret-change-me` (insecure, logs a warning) | Secret used to sign/verify JWTs |
| `PORT`       | No       | `8080`                    | Port the HTTP server listens on   |

## Deployment


- **Render** (free web service, Docker or native Go runtime)

General steps (Render example):
1. Push this repository to GitHub.
2. Create a new "Web Service" on Render, point it at the repo, and choose
   "Docker" as the environment (it will pick up the `Dockerfile`).
3. Set the `JWT_SECRET` environment variable in the dashboard.
4. Render assigns the container a `PORT` env var automatically; this app
   already reads `PORT` from the environment, so no code changes are needed.
5. Once deployed, the public health check is available at
   `https://<your-service>.onrender.com/health`.

_Deployed URL and public health check URL should be filled in here once
deployed, e.g.:_
- Deployed application URL: `<>`
- Public health check URL: `<>`
- GitHub repository link: `<>`

## Assumptions

- No admin role, ticket assignment, or comments module, per the assignment
  scope.
- A user is identified only by `username` + `password` (no email field),
  since the contract doesn't require one.
- Ticket titles are required and trimmed of surrounding whitespace;
  descriptions are optional and may be empty.
- Passwords must be at least 6 characters (a minimal sanity check, not
  specified explicitly by the contract).
- JWTs are valid for 24 hours from issuance.
- A ticket transition to its own current status (e.g. `open` -> `open`) is
  treated as invalid and rejected with `400`, to keep transition handling
  unambiguous.
- Ownership violations (ticket exists but belongs to someone else) return
  `403`; a genuinely nonexistent ticket ID returns `404`.

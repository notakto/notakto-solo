# Notakto Solo

Backend server for **Notakto** — a misere tic-tac-toe game where a player competes against an AI. In Notakto, both players place the same mark (X) on shared boards, and the player who completes a line **loses**.

## Game Rules

- One or more boards are in play simultaneously (1–5 boards).
- Both the player and AI place X on any empty cell of any live board.
- A board is **dead** when any row, column, or diagonal is fully filled.
- The player who kills the **last** remaining board **loses**.
- The AI uses parity-based strategy scaled by difficulty (1–5).

## Tech Stack

| Component       | Technology                          |
|-----------------|-------------------------------------|
| Language        | Go 1.24                             |
| HTTP Framework  | [Echo v4](https://echo.labstack.com/) |
| Database        | PostgreSQL (via [pgx](https://github.com/jackc/pgx) + [sqlc](https://sqlc.dev/)) |
| Auth            | Firebase Authentication              |
| Distributed Lock| Redis / Valkey                      |
| CI/CD           | GitHub Actions                      |

## Project Structure

```
.
├── main.go              # Entry point — server setup, DB/Redis/Firebase init
├── config/              # Environment config and game defaults
├── routes/              # Route registration
├── middleware/           # CORS, Firebase auth, per-user distributed lock
├── handlers/            # HTTP handlers (request/response binding)
├── usecase/             # Business logic (transactions, validations)
├── store/               # Database access layer (thin wrappers over sqlc)
├── logic/               # Game logic (AI moves, board checks, rewards)
├── contextkey/          # Type-safe context keys
├── db/
│   ├── migrations/      # SQL migrations (Goose)
│   ├── queries/         # SQL queries (sqlc input)
│   └── generated/       # Auto-generated Go code from sqlc
└── docs/                # Architecture and NFR docs
```

## API Endpoints

All game endpoints require a Firebase `Authorization: Bearer <token>` header.

| Method | Endpoint           | Auth | Description                          |
|--------|--------------------|------|--------------------------------------|
| POST   | `/v1/sign-in`      | Yes  | Sign in or create a new account      |
| POST   | `/v1/create-game`  | Yes  | Start a new game or resume existing  |
| POST   | `/v1/make-move`    | Yes  | Place a mark on a board cell         |
| POST   | `/v1/skip-move`    | Yes  | Pay 200 coins to skip your turn      |
| POST   | `/v1/undo-move`    | Yes  | Pay 100 coins to undo the last move  |
| POST   | `/v1/quit-game`    | Yes  | Forfeit the current game             |
| GET    | `/v1/get-wallet`   | Yes  | Get current coins and XP balance     |
| POST   | `/v1/update-name`  | Yes  | Update display name                  |
| HEAD   | `/v1/health-head`  | No   | Health check (no body)               |
| GET    | `/v1/health-get`   | No   | Health check (JSON response)         |

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL
- Redis or Valkey
- Firebase project with Authentication enabled

### Environment Variables

Create a `.env` file in the project root:

```env
PORT=1323
DATABASE_URL=postgres://user:password@localhost:5432/notakto
VALKEY_URL=redis://localhost:6379
FIREBASE_CREDENTIALS_JSON=<your Firebase service account JSON>
```

### Run

```bash
go mod download
go run main.go
```

### Run Migrations

Migrations are managed with [Goose](https://github.com/pressly/goose):

```bash
goose -dir db/migrations postgres "$DATABASE_URL" up
```

### Regenerate SQL Code

```bash
sqlc generate
```

## Architecture

```
Request → CORS → Firebase Auth → UID Lock → Handler → Usecase → Store → PostgreSQL
                                    ↑
                                  Valkey
```

- **Firebase Auth Middleware** — verifies JWT, injects UID into request context.
- **UID Lock Middleware** — acquires a per-user distributed lock via Redis/Valkey to prevent concurrent mutations.
- **Usecase Layer** — runs business logic inside serializable Postgres transactions.
- **Store Layer** — thin wrappers over sqlc-generated queries with slow-query logging (>2s).

## Game State Encoding

All moves across all boards are stored as a flat `int32[]` array. A move at board `b`, cell `c` (on a board of size `s`) is encoded as index `b * s² + c`. A parallel `bool[]` tracks whether each move was made by the AI. This makes the game history an append-only log.

## Rewards

| Outcome     | Coins                     | XP                        |
|-------------|---------------------------|---------------------------|
| Player wins | `difficulty × boards × size × rand(1–5)` | `difficulty × boards × size × rand(6–10)` |
| Player loses| 0                         | `difficulty × boards × size` (flat) |

## License

[MIT](LICENSE)

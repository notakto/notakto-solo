# Notakto Solo

Backend server for **Notakto** — a misere tic-tac-toe game where a player competes against an AI. In Notakto, both players place the same mark (X) on shared boards, and the player who completes a line **loses**.

## Game Rules

- One or more boards are in play simultaneously (1–5 boards).
- Both the player and AI place X on any empty cell of any live board.
- A board is **dead** when any row, column, or diagonal is fully filled.
- The player who kills the **last** remaining board **loses**.
- The AI uses parity-based strategy scaled by difficulty (1–5).

## Tech Stack

| Component       | Technology                         |
|-----------------|------------------------------------|
| Language        | Go 1.24                            |
| HTTP Framework  | [Echo v4](https://echo.labstack.com/) |
| Database        | PostgreSQL (via [pgx](https://github.com/jackc/pgx) + [sqlc](https://sqlc.dev/)) |
| Auth            | Firebase Authentication             |
| Rate Limiting   | Redis / Valkey (IP + UID) |
| Distributed Lock| Redis / Valkey                     |
| CI/CD           | GitHub Actions                     |

## Project Structure

```
.
├── main.go              # Entry point — server setup, DB/Redis/Firebase init
├── config/              # Environment config and game defaults
├── routes/              # Route registration
├── middleware/           # CORS, rate limiting, Firebase auth, per-user lock
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

All game and payment endpoints require a Firebase `Authorization: Bearer <token>` header. Health checks and the payment-provider webhook do not.

| Method | Endpoint                     | Auth | Description                         |
|--------|------------------------------|------|-------------------------------------|
| POST   | `/v1/sign-in`                | Yes  | Sign in or create a new account     |
| POST   | `/v1/create-game`            | Yes  | Start a new game or resume existing |
| POST   | `/v1/make-move`              | Yes  | Place a mark on a board cell        |
| POST   | `/v1/skip-move`              | Yes  | Pay 200 coins to skip your turn     |
| POST   | `/v1/undo-move`              | Yes  | Pay 100 coins to undo the last move |
| POST   | `/v1/quit-game`              | Yes  | Forfeit the current game            |
| GET    | `/v1/get-wallet`             | Yes  | Get current coins and XP balance    |
| GET    | `/v1/leaderboard`            | Yes  | Get top 10 players ordered by XP    |
| POST   | `/v1/update-name`            | Yes  | Update display name                 |
| POST   | `/v1/update-username`        | Yes  | Update username                     |
| POST   | `/v1/profile-image/upload-auth` | Yes | Create a scoped ImageKit V2 upload token |
| GET    | `/v1/all-packages`           | Yes  | List purchasable packages           |
| POST   | `/v1/create-charge`          | Yes  | Create a hosted payment charge      |
| GET    | `/v1/payment-status`         | Yes  | Get the status of a payment charge  |
| POST   | `/v1/nowpayments-webhook`    | No   | Payment-provider webhook            |
| HEAD   | `/v1/health-head`            | No   | Health check (no body)              |
| GET    | `/v1/health-get`             | No   | Health check (JSON response)        |

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
IMAGEKIT_PUBLIC_KEY=<your ImageKit public key>
IMAGEKIT_PRIVATE_KEY=<your ImageKit private key>
IMAGEKIT_URL_ENDPOINT=https://ik.imagekit.io/<your_imagekit_id>
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
Request → CORS → IP Rate Limit → Firebase Auth → UID Rate Limit → UID Lock → Handler → Usecase → Store → PostgreSQL
                      ↑                                ↑              ↑
                    Valkey ─────────────────────────────┘──────────────┘
```


- **IP Rate Limit Middleware** — sliding-window rate limit per IP via Redis/Valkey (120 req window).
- **Firebase Auth Middleware** — verifies JWT, injects UID into request context.
- **UID Rate Limit Middleware** — sliding-window rate limit per authenticated UID via Redis/Valkey (60 req window).
- **UID Lock Middleware** — acquires a per-user distributed lock via Redis/Valkey to prevent concurrent mutations.
- **Usecase Layer** — runs business logic inside serializable Postgres transactions.
- **Store Layer** — thin wrappers over sqlc-generated queries with slow-query logging (>2s).

### Profile images

Profile images use a direct-to-ImageKit upload flow. The authenticated frontend first calls
`POST /v1/profile-image/upload-auth` with the original filename:

```json
{"fileName":"avatar.webp"}
```

The response contains a five-minute ImageKit Upload V2 JWT, the V2 upload URL, and an
`uploadPayload`. The frontend must append every returned payload field unchanged to a
multipart form along with `file` and `token`, then send it directly to ImageKit. Upload V2
uses raw multipart requests; the current browser SDK upload helper uses the V1 protocol and
is not compatible with this token. The server fixes the destination to
`/profile-images/<base64url-uid>/avatar-<uuid>.<ext>` and signs the filename, folder,
overwrite/publication flags, and checks into the token; clients cannot choose another path.

```json
{
  "token": "<one-time-jwt>",
  "expiresAt": 1234567890,
  "uploadUrl": "https://upload.imagekit.io/api/v2/files/upload",
  "uploadPayload": {
    "fileName": "avatar-<uuid>.webp",
    "folder": "/profile-images/<encoded-uid>",
    "useUniqueFileName": "false",
    "overwriteFile": "false",
    "isPrivateFile": "false",
    "isPublished": "true",
    "checks": "<signed checks expression>"
  }
}
```

### Leaderboard

`GET /v1/leaderboard` returns up to 10 players with non-null XP, ranked by competition ranking (`1, 1, 3`). Ties are ordered alphabetically by username and only `rank`, `username`, and `xp` are exposed.

```json
{
  "leaderboard": [
    {"rank": 1, "username": "player_one", "xp": 1200}
  ]
}
```

## Game State Encoding

All moves across all boards are stored as a flat `int32[]` array. A move at board `b`, cell `c` (on a board of size `s`) is encoded as index `b * s² + c`. A parallel `bool[]` tracks whether each move was made by the AI. This makes the game history an append-only log.

## Rewards

| Outcome     | Coins                     | XP                        |
|-------------|---------------------------|---------------------------|
| Player wins | `difficulty × boards × size × rand(1–5)` | `difficulty × boards × size × rand(6–10)` |
| Player loses| 0                         | `difficulty × boards × size` (flat) |

## License

[MIT](LICENSE)

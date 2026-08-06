# Commerce Lab

Next.js + Go + PostgreSQL の EC サイトサンプルです。

## Stack

- Frontend: Next.js App Router, TypeScript, Tailwind CSS, TanStack Query, Zod
- Backend: Go, chi, pgxpool, zerolog
- DB: PostgreSQL
- Seed: Go command with batched inserts and PostgreSQL `COPY`

Node.js is pinned by Volta in `frontend/package.json`: Node `22.17.0`, npm `10.9.2`.

## Quick Start

```bash
cp .env.example .env
docker compose up -d postgres
./scripts/migrate.sh
./scripts/seed.sh
docker compose up backend
cd frontend && npm install && npm run dev
```

Or run the application stack with Docker:

```bash
docker compose up --build
```

Frontend: http://localhost:3000
Backend: http://localhost:8080
Object storage: http://localhost:9001
Auth server: http://localhost:18080
Frontend login: http://localhost:3000/login

## Container Images

Build the production images for ECS/ECR:

```bash
docker build -t commerce-lab-backend:latest ./backend
docker build \
  --build-arg NEXT_PUBLIC_API_BASE_URL=https://api.example.com \
  --build-arg NEXT_PUBLIC_AUTH_ISSUER=https://auth.example.com/realms/commerce \
  --build-arg NEXT_PUBLIC_AUTH_CLIENT_ID=commerce-frontend \
  -t commerce-lab-frontend:latest ./frontend
```

The frontend `NEXT_PUBLIC_*` values are compiled into the browser bundle, so set them to the target environment URLs at image build time.

MinIO login:

- User: `minio`
- Password: `minio123`

Product images are stored in the `commerce-images` bucket and seeded under `products/`.

Keycloak login:

- Admin console user: `admin@example.test`
- Admin console password: `admin123`
- Bootstrap admin user: `admin` / `admin123`
- Realm: `commerce`
- Demo customer: `customer@example.test` / `customer123`
- Demo EC admin in `commerce` realm: `admin@example.test` / `admin123`

If local port `5432` is already used:

```bash
POSTGRES_PORT=55432 docker compose up -d postgres
DATABASE_URL=postgres://commerce:commerce@localhost:55432/commerce?sslmode=disable ./scripts/migrate.sh
DATABASE_URL=postgres://commerce:commerce@localhost:55432/commerce?sslmode=disable ./scripts/seed.sh
```

## Seed Scale

Default scale is `dev`.

```bash
SEED_SCALE=30k ./scripts/seed.sh
SEED_SCALE=100k ./scripts/seed.sh
SEED_SCALE=dev ./scripts/seed.sh
SEED_SCALE=medium ./scripts/seed.sh
SEED_SCALE=large ./scripts/seed.sh
```

Approximate generated rows:

| Scale | Products | Customers | Admins | Orders | Order items | Auth users | Status events | Shipments | Shipment events | Admin notes | Reviews |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 30k | 30,000 | 30,000 | 30,000 | 30,000 | 30,000 | 60,000 | 30,000 | 30,000 | 30,000 | 30,000 | 30,000 |
| 100k | 100,000 | 100,000 | 100,000 | 100,000 | 100,000 | 200,000 | 100,000 | 100,000 | 100,000 | 100,000 | 100,000 |
| dev | 20,000 | 50,000 | 200 | 100,000 | 300,000 | 50,200 | 450,000 | 90,000 | 270,000 | 40,000 | 80,000 |
| medium | 200,000 | 500,000 | 2,000 | 1,000,000 | 3,000,000 | 502,000 | 4,500,000 | 900,000 | 2,700,000 | 400,000 | 800,000 |
| large | 1,000,000 | 2,000,000 | 10,000 | 5,000,000 | 15,000,000 | 2,010,000 | 22,500,000 | 4,500,000 | 13,500,000 | 2,000,000 | 4,000,000 |

`large` はパフォーマンスチューニング検証用です。DB サイズと投入時間が大きくなるため、ローカルでは `medium` から試してください。

`30k` と `100k` は各 seed 対象データを同じ件数に揃える検証用プリセットです。顧客アカウントと EC 管理者アカウントをそれぞれ作るため、`auth_users` は合計で指定件数の 2 倍になります。

## Useful Commands

```bash
docker compose up -d postgres
docker compose up backend
cd frontend && npm run dev
cd backend && go test ./...
```

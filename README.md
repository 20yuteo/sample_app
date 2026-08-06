# Commerce Lab

Next.js + Go + PostgreSQL の EC サイトサンプルです。

## 技術スタック

- フロントエンド: Next.js App Router, TypeScript, Tailwind CSS, TanStack Query, Zod
- バックエンド: Go, chi, pgxpool, zerolog
- DB: PostgreSQL
- Seed: Go コマンドによるバッチ insert と PostgreSQL `COPY`

Node.js は `frontend/package.json` の Volta 設定で固定しています: Node `22.17.0`, npm `10.9.2`。

## クイックスタート

```bash
cp .env.example .env
docker compose up -d postgres
./scripts/migrate.sh
./scripts/seed.sh
docker compose up backend
cd frontend && npm install && npm run dev
```

Docker だけでアプリケーション全体を起動する場合:

```bash
docker compose up --build
```

フロントエンド: http://localhost:3000
バックエンド: http://localhost:8080
オブジェクトストレージ: http://localhost:9001
認証サーバー: http://localhost:18080
フロントエンドログイン: http://localhost:3000/login

## コンテナイメージ

ECS/ECR 向けの本番用イメージをビルドします:

```bash
docker build -t commerce-lab-backend:latest ./backend
docker build \
  --build-arg NEXT_PUBLIC_API_BASE_URL=https://api.example.com \
  --build-arg NEXT_PUBLIC_AUTH_ISSUER=https://auth.example.com/realms/commerce \
  --build-arg NEXT_PUBLIC_AUTH_CLIENT_ID=commerce-frontend \
  -t commerce-lab-frontend:latest ./frontend
```

フロントエンドの `NEXT_PUBLIC_*` はブラウザ向けバンドルに埋め込まれるため、イメージのビルド時にデプロイ先環境のURLを指定してください。

ECS/Fargate のデプロイテンプレート、GitHub Actions workflow、ECR push 用スクリプトは `deploy/aws/` と `.github/workflows/` にあります。

MinIO ログイン:

- ユーザー: `minio`
- パスワード: `minio123`

商品画像は `commerce-images` バケットに保存され、`products/` 配下に seed されます。

Keycloak ログイン:

- 管理コンソールユーザー: `admin@example.test`
- 管理コンソールパスワード: `admin123`
- ブートストラップ管理ユーザー: `admin` / `admin123`
- Realm: `commerce`
- デモ顧客: `customer@example.test` / `customer123`
- `commerce` realm のデモ EC 管理者: `admin@example.test` / `admin123`

ローカルの `5432` ポートがすでに使われている場合:

```bash
POSTGRES_PORT=55432 docker compose up -d postgres
DATABASE_URL=postgres://commerce:commerce@localhost:55432/commerce?sslmode=disable ./scripts/migrate.sh
DATABASE_URL=postgres://commerce:commerce@localhost:55432/commerce?sslmode=disable ./scripts/seed.sh
```

## Seed 規模

デフォルトの規模は `dev` です。

```bash
SEED_SCALE=30k ./scripts/seed.sh
SEED_SCALE=100k ./scripts/seed.sh
SEED_SCALE=dev ./scripts/seed.sh
SEED_SCALE=medium ./scripts/seed.sh
SEED_SCALE=large ./scripts/seed.sh
```

生成されるおおよその行数:

| 規模 | 商品 | 顧客 | 管理者 | 注文 | 注文明細 | 認証ユーザー | ステータスイベント | 配送 | 配送イベント | 管理メモ | レビュー |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 30k | 30,000 | 30,000 | 30,000 | 30,000 | 30,000 | 60,000 | 30,000 | 30,000 | 30,000 | 30,000 | 30,000 |
| 100k | 100,000 | 100,000 | 100,000 | 100,000 | 100,000 | 200,000 | 100,000 | 100,000 | 100,000 | 100,000 | 100,000 |
| dev | 20,000 | 50,000 | 200 | 100,000 | 300,000 | 50,200 | 450,000 | 90,000 | 270,000 | 40,000 | 80,000 |
| medium | 200,000 | 500,000 | 2,000 | 1,000,000 | 3,000,000 | 502,000 | 4,500,000 | 900,000 | 2,700,000 | 400,000 | 800,000 |
| large | 1,000,000 | 2,000,000 | 10,000 | 5,000,000 | 15,000,000 | 2,010,000 | 22,500,000 | 4,500,000 | 13,500,000 | 2,000,000 | 4,000,000 |

`large` はパフォーマンスチューニング検証用です。DB サイズと投入時間が大きくなるため、ローカルでは `medium` から試してください。

`30k` と `100k` は各 seed 対象データを同じ件数に揃える検証用プリセットです。顧客アカウントと EC 管理者アカウントをそれぞれ作るため、`auth_users` は合計で指定件数の 2 倍になります。

## よく使うコマンド

```bash
docker compose up -d postgres
docker compose up backend
cd frontend && npm run dev
cd backend && go test ./...
```

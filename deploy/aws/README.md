# AWS ECS デプロイ

このディレクトリには、アプリケーションを ECS Fargate で動かすためのデプロイテンプレートを置いています。

## 前提

- バックエンドとフロントエンドは別々の ECS サービスとして動かします。
- PostgreSQL は ECS の外で動かします。通常は RDS を使います。
- `DATABASE_URL` は AWS Secrets Manager に保存します。
- 外部からのトラフィックは Application Load Balancer 経由で各サービスに流します。
- フロントエンドの `NEXT_PUBLIC_*` はブラウザ向けバンドルに埋め込まれるため、イメージのビルド時に指定します。

## 必要な AWS リソース

- ECR リポジトリ2つ:
  - `commerce-lab-backend`
  - `commerce-lab-frontend`
- ECS クラスター。例: `commerce-lab`
- ECR pull と CloudWatch Logs への書き込み権限を持つ ECS task execution role
- 次の通信を許可する security group:
  - ALB からフロントエンド task の `3000` ポート
  - ALB またはフロントエンドクライアントからバックエンド task の `8080` ポート
  - バックエンド task から RDS PostgreSQL の `5432` ポート
- CloudWatch log group:
  - `/ecs/commerce-lab-backend`
  - `/ecs/commerce-lab-frontend`

## GitHub Actions でのビルドとプッシュ

`main` ブランチに push されると、`.github/workflows/build-and-push-images.yml` が backend/frontend の Docker イメージをビルドし、ECR に push します。

GitHub リポジトリに次の Variables を設定してください:

- `AWS_ACCOUNT_ID`: AWSアカウントID。例: `123456789012`
- `AWS_REGION`: AWSリージョン。例: `ap-northeast-1`
- `BACKEND_PUBLIC_URL`: ブラウザから見えるバックエンドURL。例: `https://api.example.com`
- `AUTH_ISSUER`: 認証サーバーのIssuer URL。例: `https://auth.example.com/realms/commerce`
- `AUTH_CLIENT_ID`: フロントエンド用KeycloakクライアントID。未設定の場合は `commerce-frontend`

GitHub リポジトリに次の Secret を設定してください:

- `AWS_ROLE_ARN`: GitHub Actions から assume role する IAM role ARN

AWS側では、GitHub OIDC provider と、対象リポジトリの `main` ブランチから assume role できる IAM role を作成しておきます。その role には少なくとも ECR への push 権限が必要です。

push されるイメージタグはコミットSHAの先頭7文字です。

## ローカルから手動でビルドとプッシュする場合

通常の運用では GitHub Actions から実行します。ローカルからの実行は初期検証や手動確認用です。

AWSアカウントとデプロイ先環境に合わせて、次の値を設定します:

```bash
export AWS_ACCOUNT_ID=123456789012
export AWS_REGION=ap-northeast-1
export IMAGE_TAG=$(git rev-parse --short HEAD)
export FRONTEND_PUBLIC_URL=https://app.example.com
export BACKEND_PUBLIC_URL=https://api.example.com
export AUTH_ISSUER=https://auth.example.com/realms/commerce
```

ECR にログインします:

```bash
aws ecr get-login-password --region "$AWS_REGION" \
  | docker login --username AWS --password-stdin "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"
```

ビルドして push します:

```bash
./deploy/aws/build-and-push.sh
```

## Task Definition の登録

task definition JSON のプレースホルダを実値に置き換えます:

- `<AWS_ACCOUNT_ID>`
- `<AWS_REGION>`
- `<ECS_EXECUTION_ROLE_ARN>`
- `<ECS_TASK_ROLE_ARN>`
- `<DATABASE_URL_SECRET_ARN>`
- `<FRONTEND_PUBLIC_URL>`

置き換えたら登録します:

```bash
aws ecs register-task-definition --cli-input-json file://deploy/aws/backend-task-definition.json
aws ecs register-task-definition --cli-input-json file://deploy/aws/frontend-task-definition.json
```

登録した task definition から ECS サービスを作成または更新します。target group のヘルスチェックは次のように設定します:

- バックエンド: `8080` ポートの `GET /healthz`
- フロントエンド: `3000` ポートの `GET /`

## ランタイム設定

バックエンドの環境変数:

- `API_ADDR=:8080`
- `CORS_ALLOWED_ORIGINS=https://app.example.com`
- Secrets Manager から渡す `DATABASE_URL`

フロントエンドのビルド引数:

- `NEXT_PUBLIC_API_BASE_URL=https://api.example.com`
- `NEXT_PUBLIC_AUTH_ISSUER=https://auth.example.com/realms/commerce`
- `NEXT_PUBLIC_AUTH_CLIENT_ID=commerce-frontend`

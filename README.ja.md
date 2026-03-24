# Fluent Manager

**[English](README.md)** | **[中文](README.zh.md)**

[Fluent Bit](https://fluentbit.io/) と [Fluentd](https://www.fluentd.org/) エージェントの統合管理プラットフォームです。トポロジー、設定、デプロイ、モニタリングをひとつのダッシュボードで管理できます。

## 主な機能

- **インフラトポロジー** — データセンター・リージョン・クラスターによるノード管理、自動振り分けルール
- **設定管理** — テンプレートベースの設定、バージョン管理、モジュール、プレビュー、Lint
- **リモートデプロイ** — ノード・クラスター・データセンター単位で設定をプッシュ配信
- **リアルタイム監視** — ハートビート、メトリクス収集、ランタイムドリフト検知、ヘルスビュー
- **ログパイプライン可視化** — ソースからアグリゲーターへのフォワーディングをグラフ表示
- **AI 分析支援** — LLM によるログサンプル分析と設定生成
- **RBAC & スコープ** — ロールベース権限 + トポロジーレベルのスコープ制御
- **マルチ認証** — ローカル / LDAP / SAML 認証
- **エージェントポリシー** — グローバル → 環境 → クラスター → ラベルセレクターの階層型オーバーライド
- **多言語対応** — 英語・中国語・日本語 UI

## デプロイ方法

### 方法 1：オールインワンバイナリ

フロントエンドを内蔵した単一バイナリ。コピーして実行するだけの最もシンプルな方法です。

```bash
make build-all-in-one            # ローカルプラットフォーム
make build-all-in-one-linux      # Linux amd64 + arm64
```

```bash
cp config.yaml.example config.yaml
# config.yaml を編集（データベース、認証など）
./fluent-manager-server
```

API と Web UI を同じポート（デフォルト `:8080`）で提供します。

### 方法 2：フロントエンド分離構成

バックエンドは API のみ。フロントエンドは静的ファイルとして Nginx 等に配置します。

```bash
# バックエンドのビルド
make build-server-linux

# フロントエンドのビルド＆パッケージ
make frontend-package            # bin/fluent-manager-frontend.tar.gz を生成
```

**バックエンド** — `config.yaml` を設定してバイナリを実行。

**フロントエンド** — tarball を Web サーバーのルートに展開：

```bash
tar -xzf fluent-manager-frontend.tar.gz -C /usr/share/nginx/html
```

Nginx 設定例：

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location /api/ {
        proxy_pass http://backend:8080;
    }
    location /saml/ {
        proxy_pass http://backend:8080;
    }
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

### 方法 3：Docker

```bash
# サーバー（フロントエンド含む）
docker build -t fluent-manager .
docker run -p 8080:8080 -v ./config.yaml:/app/config.yaml fluent-manager

# エージェントのみ
docker build --target runtime-agent -t fluent-manager-agent .
```

## エージェント

各管理対象ノードにデプロイする軽量な Go バイナリです。ハートビート、メトリクス収集、設定同期、リモートコマンド実行を行います。

```bash
make build-agent                 # ローカルプラットフォーム
make build-agent-linux           # Linux amd64 + arm64
```

`agent.yaml`（`agent.yaml.example` 参照）で設定します。必須項目は `server_url` と `api_key` のみ — その他はサーバーからエージェントポリシー経由で配信可能です。

## クイックスタート

1. サーバーを起動（上記いずれかの方法）
2. Web UI を開く — セットアップウィザードが初期設定（データベース、管理者アカウント、認証など）を案内します
3. トポロジーを構築（データセンター → リージョン → クラスター）
4. ノードにエージェントをデプロイ
5. 設定テンプレートを作成してプッシュ配信

## ライセンス

MIT

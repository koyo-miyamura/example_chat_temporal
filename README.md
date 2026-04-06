## example-chat-temporal

Temporal を使ったチャット API のサンプルコード。

ユーザーのリクエストをワークフローで管理し、2 ステップの処理を順番に実行する。

## アーキテクチャ

```text
[client] --POST /chat--> [handler] --UpdateWithStart--> [Temporal] <-- [worker]
```

- **handler** — HTTP サーバー。リクエストを受け取り Temporal にワークフローの起動・更新を依頼する
- **worker** — Temporal ワーカー。ワークフローとアクティビティを実行する
- **Temporal** — ワークフローエンジン。状態管理・タスクキューを担う

## ワークフロー概要

`ChatWorkflow` は 2 ステップで構成される。

1. ユーザーメッセージを受信 → `ChatStep1Activity` を実行（約 3 秒）
2. 再度ユーザーメッセージを受信 → `ChatStep2Activity` を実行（約 3 秒）→ ワークフロー完了

各ステップは Temporal の Update を通じてトリガーされる。アクティビティ処理中は次の Update をバリデーションで拒否する。

## 起動方法

```bash
docker compose up
```

以下のサービスが起動する。

| サービス | 役割 |
| --- | --- |
| handler | HTTP サーバー（ポート 8080） |
| worker | Temporal ワーカー |
| temporal | Temporal サーバー |
| temporal-ui | Temporal Web UI（ポート 18880） |
| mysql | データベース |

## 確認方法

### Step1 のトリガー

```bash
curl -X POST "http://localhost:8080/chat?chat_id=chat-001&user_id=user-1&data=start"
```

### Step2 のトリガー（Step1 完了後に同じ `chat_id` で送信）

```bash
curl -X POST "http://localhost:8080/chat?chat_id=chat-001&user_id=user-1&data=next"
```

Temporal UI（<http://localhost:18880>）でワークフローの状態を確認できる。

### エラーケース

処理中に送信した場合（409）:

```bash
# Step1 のアクティビティ実行中（3秒以内）に送信
curl -X POST "http://localhost:8080/chat?chat_id=chat-001&user_id=user-1&data=too-fast"
# => HTTP 409: Chat is currently processing, please wait
```

完了済みのチャットに送信した場合（409）:

```bash
# Step2 まで完了した後に同じ chat_id で送信
curl -X POST "http://localhost:8080/chat?chat_id=chat-001&user_id=user-1&data=too-late"
# => HTTP 409: Chat already completed
```

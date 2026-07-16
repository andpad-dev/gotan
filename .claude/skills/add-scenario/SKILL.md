---
name: add-scenario
description: ワークショップの問題（README 1 枚に設問・ヒント・答えを収録）を作成し、難易度ラベル付きの Draft PR まで作るスキル。問題追加、シナリオ追加、例題作成を依頼されたときに使用する。
---

# add-scenario

問題の作成から Draft PR 作成までを、リポジトリの規約（CLAUDE.md）に沿って行う。

## 前提

- リポジトリ直下の `CLAUDE.md` を読み、規約（配置・フォーマット・品質規約）を把握していること。
- `gh` で認証済みであること。

## 手順

### 1. 入力の確認

以下が揃っているか確認し、不足があれば AskUserQuestion で確認する。

- **難易度**: `初級`（01-beginner）/ `中級`（02-intermediate）/ `上級`（03-advanced）
- **テーマ**: 何を調べさせる問題か（題材のコード・リンクがあれば受け取る）
- **カテゴリ**: `01-packages` / `02-features` / `03-cmd-tools` / `04-deep-dive`
  - テーマから自明な場合は確認せず選んでよい

### 2. 例題の確認

該当する難易度の例題 PR（初級 #2 / 中級 #3 / 上級 #4）の diff を `ghro pr diff` で確認し、構成・文体・分量の基準にする。
マージ済みなら該当ディレクトリの実ファイルを読む。

### 3. 裏取り（ファイル作成より先に行う）

- 参照する URL はすべて `curl -s -o /dev/null -w "%{http_code}"` で実在を確認する。
- 題材のコードは scratchpad で `go run` し、実際の出力を取得する。
  - 渡されたコードが動かない場合は、勝手に直さず修正案と理由をユーザーに報告して判断を仰ぐ。
- 問題コード・実行例には Go Playground の共有リンクを作る:
  ```bash
  curl -s -X POST --data-binary @main.go https://play.golang.org/share
  # 返ってきた ID を https://go.dev/play/p/<ID> にする。作成後 200 を確認する。
  ```
- 解説に書く仕様・設計の説明は一次情報（pkg.go.dev / go.dev/ref/spec / Go 本体ソース）から該当文言を確認する。

### 4. ファイル作成

- `<カテゴリ>/<難易度>/<問題名>/README.md` を 1 枚作成する。
- 構成は CLAUDE.md の「問題 README の構成」に従う（ヒントと答えは `<details>` で折りたたむ）。
- 設問数は難易度の目安（初級 2 つほど / 中級 3 つ / 上級 制限なし）に合わせる。
- 「まず `f` キーで検索する」のような共通の初手はヒントに書かず、カテゴリルートの README（逆引き手順）に載っているか確認する。カテゴリ README が無ければ作成を提案する。

### 5. 自己検証

PR テンプレートのチェックリストを 1 項目ずつ確認する。
確認できていない項目があれば手順 3 に戻る。

### 6. Draft PR 作成

```bash
git fetch origin
git switch -c feature/scenario-<問題名> origin/main
git add <カテゴリ>/<難易度>/<問題名>/README.md
git commit -m "feat: <難易度>問題「<タイトル>」を追加"
git push -u origin feature/scenario-<問題名>
gh pr create --draft --base main --label "<難易度>" --title "<難易度>: <タイトル>"
```

- ラベルは配置ディレクトリの難易度と一致させる（CI が検証する）。
- PR body はテンプレートに沿って書く（概要 / 問題のテーマ / チェックリスト）。
- 案出し時のメモや経緯は PR body に書かない。
- 元ネタから修正した点（動かないコードの修正など）があれば「補足」に明記する。

### 7. 報告

作成した PR の URL、設問の構成、裏取りで確認した内容（修正点があればそれも）をユーザーに報告する。
Ready 化はユーザーの判断に委ねる。

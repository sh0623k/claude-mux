# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

claude-muxは複数のClaudeエージェントを自動的に管理するためのGo製CLIアプリケーションです。tmuxセッションを作成し、複数のペインに分割してClaudeエージェントを起動する機能を提供します。

## Development Commands

### Building
```bash
go build -o claude-mux cmd/claudemux/main.go
```

### Testing
```bash
go test ./...
```

### Running
```bash
./claude-mux
```

### Module Management
```bash
go mod tidy
go mod download
```

## Architecture

### Clean Architecture Structure
- `cmd/claudemux/main.go`: エントリーポイント、tmuxセッション管理のメインロジック
- `pkg/gateways/tmux/`: tmux操作を抽象化するインターフェースと実装
  - `gateway.go`: Gatewayインターフェースと具象実装
  - `command_factory.go`: tmuxコマンド生成の責務を分離

### Core Components
- `Gateway` interface: tmux操作の抽象化レイヤー
  - セッション作成/削除/アタッチ
  - ペイン分割とレイアウト設定（水平分割、main-verticalレイアウト）
  - ペインサイズ調整（メインペイン25%幅、サブペイン25%高さ）
  - Claudeエージェント起動と対話モード設定
- `commandFactory`: tmuxコマンド生成の責務分離
- Signal handling: SIGINT/SIGTERMによる終了時クリーンアップ処理

### Dependencies
- Go 1.23を使用
- 外部依存関係なし（標準ライブラリのみ使用）

## File Structure
```
├── cmd/claudemux/main.go          # メインエントリーポイント
├── pkg/gateways/tmux/
│   ├── gateway.go                 # tmux操作のインターフェースと実装
│   └── command_factory.go         # tmuxコマンド生成ファクトリ
├── go.mod                         # モジュール定義
└── claude-mux                     # ビルド済みバイナリ
```

## Key Implementation Details

### Session Management
- デフォルトセッション名: "manager-session"
- 終了時に自動的にセッションをクリーンアップ
- シグナルハンドリングによる適切な終了処理

### Pane Configuration
- メインペイン: 左端（コマンド実行用、25%幅）、対話モード有効
- サブペイン: 右側に最大4分割（Claudeエージェント用、各25%高さ）
- レイアウト: main-vertical
- 分割制限: 最大4ペインまで（SplitWindowHorizontally関数で制限）

### Agent Startup
- 左端以外のペインでClaudeエージェント（`claude`コマンド）を自動起動
- ペインの準備完了を待機（プロンプト文字検出による）してからコマンド実行
- メインペインの対話モードにより、入力した内容を全サブペインに同時送信
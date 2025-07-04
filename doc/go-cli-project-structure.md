# Go CLIアプリケーション構成ガイド

GoのCLIアプリケーション構成のベストプラクティスガイド。

## ディレクトリ構成

```
myapp/                       # Goモジュールルート (go.mod)
├── cmd/                     # バイナリごとのフォルダ
│   └── myapp/               # CLIのmainパッケージ
│       ├── main.go          # エントリーポイント
│       ├── root.go          # CLIの配線とフラグ設定
│       ├── serve.go         # サブコマンド
│       └── completion.go    # シェル補完
├── internal/                # プライベートアプリケーションコード
│   ├── config/              # Viperベースの設定ローダー
│   ├── service/             # ビジネスロジック
│   └── testdata/            # go testで無視されるヘルパーデータ
├── pkg/                     # 再利用可能なライブラリ (パブリックAPI)
│   ├── entities/            # ドメインエンティティ
│   ├── gateways/            # 外部サービスとのインタフェース
│   ├── usecases/            # ユースケース層
│   ├── interfaces/          # インタフェース層
│   └── presenters/          # プレゼンテーション層
├── scripts/                 # ヘルパースクリプト (lint, release)
├── build/                   # Docker/パッケージ固有ファイル
├── assets/                  # 静的ファイル (埋め込み用)
├── docs/                    # 追加のMarkdown/manページ
├── Makefile                 # 便利なターゲット
├── go.mod                   # Goモジュール定義
├── go.sum                   # 依存関係のハッシュ
└── .github/workflows/       # CI (lint, vet, test, go run ./...)
```

## 設計原則

### クリーンアーキテクチャの適用

`pkg/`配下の構成は、クリーンアーキテクチャに基づいている：

- `entities/`: ドメインオブジェクトとビジネスルール
- `usecases/`: アプリケーションのユースケース
- `gateways/`: 外部サービスとのインタフェース
- `interfaces/`: UIレイヤー
- `presenters/`: プレゼンテーション層

### 依存関係の方向

```
cmd/ → internal/ ← pkg/
         ↓
    entities ← usecases ← gateways
         ↑       ↑
   interfaces  presenters
```

## 主要コンポーネント

### 1. エントリーポイント (`cmd/`)

#### `cmd/myapp/main.go`
```go
package main

import "github.com/you/myapp/cmd/myapp"

func main() {
    myapp.Execute()
}
```

#### `cmd/myapp/root.go`
```go
package main

import (
    "github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
    Use:           "myapp",
    Short:         "CLIアプリケーションの説明",
    SilenceUsage:  true,   // エラー時にusageを表示しない
    SilenceErrors: true,   // main()で適切にエラーを表示
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}

func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "設定ファイルパス")
}
```

### 2. 設定管理 (`internal/config/`)

ViperとCobraを組み合わせた設定管理：

```go
func Load(cmd *cobra.Command) (*Config, error) {
    v := viper.New()
    v.SetConfigName("config")
    v.AddConfigPath("$HOME/.myapp")
    v.SetEnvPrefix("MYAPP")
    v.AutomaticEnv()
    
    if cfgFile, _ := cmd.Flags().GetString("config"); cfgFile != "" {
        v.SetConfigFile(cfgFile)
    }
    
    // 設定の読み込みとバインド
}
```

### 3. ドメイン層 (`pkg/entities/`)

#### 構成例
- `formatters/`: 出力フォーマット処理
- `config/`: 設定関連エンティティ

### 4. ゲートウェイ層 (`pkg/gateways/`)

外部サービスとのインタフェース：
- `api/`: APIクライアント
- `io/file/`: ファイルシステムアクセス

### 5. ユースケース層 (`pkg/usecases/`)

ビジネスロジックの実装：
- `data_fetcher.go`: データ取得
- `processor.go`: データ処理
- `factory.go`: ファクトリーパターン

## 開発時のベストプラクティス

### 1. テスト構成

- 各パッケージにテストファイルを配置
- `testdata/`ディレクトリにテスト用データを格納
- モックはファイルに集約

### 2. 依存関係の管理

```bash
# Goモジュールの初期化
go mod init github.com/your-org/your-app

# 主要な依存関係
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get github.com/stretchr/testify@latest
```

### 3. ビルドとリリース

#### Makefile例
```makefile
.PHONY: build test lint clean

APP_NAME := myapp
VERSION := $(shell git describe --tags --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
```

### 4. シェル補完の実装

```go
var completionCmd = &cobra.Command{
    Use:   "completion [bash|zsh|fish|powershell]",
    Short: "シェル補完スクリプトを生成",
    RunE: func(cmd *cobra.Command, args []string) error {
        switch args[0] {
        case "bash":
            return cmd.Root().GenBashCompletion(os.Stdout)
        case "zsh":
            return cmd.Root().GenZshCompletion(os.Stdout)
        // 他のシェル対応
        }
    },
}
```

## 実装チェックリスト

### 基本構成
- [ ] go.modの設定
- [ ] cmd/配下のエントリーポイント
- [ ] internal/配下のプライベートコード
- [ ] pkg/配下のパブリックAPI

### CLI機能
- [ ] Cobraベースのコマンド構成
- [ ] Viperによる設定管理
- [ ] フラグとサブコマンドの実装
- [ ] シェル補完機能

### アーキテクチャ
- [ ] クリーンアーキテクチャの適用
- [ ] 依存関係の適切な方向
- [ ] インタフェースによる疎結合

### 品質保証
- [ ] ユニットテストの実装
- [ ] CI/CDパイプライン
- [ ] リンターの設定
- [ ] ドキュメント整備

この構成に従うことで、保守性が高く、テストしやすい Go CLIアプリケーションが作成できる。
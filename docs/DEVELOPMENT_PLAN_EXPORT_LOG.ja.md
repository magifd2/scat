## 開発計画: チャネルログエクスポート機能 (最終版)

### I. 機能概要

この機能により、ユーザーは指定されたSlackチャネルからチャットログをエクスポートできます。エクスポートされたデータは構造化され（JSON形式）、ユーザーIDと解決されたユーザー名の両方を含み、オプションで添付ファイルをローカルディレクトリにダウンロードします。

### II. コア原則の遵守

*   **セキュリティ第一:** すべてのSlack APIトークンは安全に処理されます。ファイルのダウンロードは検証されます。**パス・トラバーサル、インジェクション、その他の脆弱性を防ぐために、データ検証とサニタイズが厳密に適用されます。**
*   **テスト容易性第一:** 新しい関数とコンポーネントは単体テストを念頭に置いて設計されます。
*   **明示的であること:** 依存関係は明示的に渡されます。
*   **コードスタイルと品質:** 標準のGoフォーマットと慣用的なGoプラクティスを遵守します。
*   **依存関係管理:** Go Modulesを使用します。新しい依存関係を追加した後には`go mod tidy`を実行します。

### III. 「大規模改修」の定義

この機能は新しい機能性を導入するものであり、初期化ロジック、コアインターフェース（`Provider`など）、または3つ以上の既存パッケージにわたる広範な影響を伴う変更は含まれません。したがって、`GEMINI.md` III-1に従い、「大規模改修」とは見なされません。

### IV. 安全なリファクタリングプロトコル

開発は、最小限の変更を行い、すぐにテストし、成功したらコミットするという安全なリファクタリングプロトコル（III-2）に従います。

### V. 詳細計画

#### フェーズ1：プロバイダーインターフェースとAPI統合

1.  **`Capabilities`構造体の更新:**
    *   **ファイル:** `internal/provider/provider.go`
    *   **変更点:** `Capabilities`構造体に新しいブール値フィールドを追加し、既存のコードとの一貫性を保ちます。
        ```go
        type Capabilities struct {
            // ... 既存のフィールド
            CanExportLogs bool
        }
        ```

2.  **新しい`LogExporter`インターフェースの定義:**
    *   **ファイル:** `internal/provider/provider.go`
    *   **変更点:** エクスポート機能専用の新しいインターフェースを作成し、関心を分離します。
        ```go
        type LogExporter interface {
            GetConversationHistory(channelID string, latest, oldest string, limit int, cursor string) (*ConversationHistoryResponse, error)
            GetUserInfo(userID string) (*UserInfoResponse, error)
            DownloadFile(fileURL string) ([]byte, error)
        }
        ```
    *   メインの`provider.Interface`には、このエクスポーターを取得するメソッドを追加します。
        ```go
        type Interface interface {
            Capabilities() Capabilities
            // ... 既存のメソッド
            LogExporter() LogExporter // CanExportLogsがtrueの場合に呼び出される
        }
        ```

3.  **プロバイダレベルのAPIデータ構造の定義:**
    *   **ファイル:** `internal/provider/types.go` (新規ファイル)
    *   **変更点:** `LogExporter`インターフェースによって返される、プロバイダに依存しない共通のAPI応答構造体を保持するためにこのファイルを作成します。
        *   `ConversationHistoryResponse`
        *   `UserInfoResponse`

4.  **Slackプロバイダーの実装:**
    *   **ファイル:** `internal/provider/slack/slack.go` および `internal/provider/slack/api.go`
    *   **変更点:**
        *   `Capabilities()`メソッドを更新し、`CanExportLogs: true`を返します。
        *   `LogExporter()`メソッドを実装し、`LogExporter`インターフェースを満たす具象型を返します。
        *   `GetConversationHistory`、`GetUserInfo`、`DownloadFile`メソッドを実装します。

5.  **モックプロバイダーの更新:**
    *   **ファイル:** `internal/provider/mock/mock.go`
    *   **変更点:**
        *   `Capabilities()`メソッドを更新し、`CanExportLogs: true`（テストのため）を返します。
        *   `LogExporter()`メソッドを実装し、モック版の`LogExporter`を返します。

#### フェーズ2：コアロジックとデータ構造

1.  **汎用エクスポートデータモデルの定義:**
    *   **ファイル:** `internal/export/types.go` (新規ファイル)
    *   **構造体:** `ExportedLog`, `ExportedMessage`, `ExportedFile`など、最終的な出力形式に合わせた汎用的な構造体を定義します。

2.  **コア`Exporter`ロジックの実装:**
    *   **ファイル:** `internal/export/exporter.go` (新規ファイル)
    *   **変更点:** エクスポートプロセス全体を調整する中央の`Exporter`構造体を作成します。コンストラクタで`provider.LogExporter`を受け取ります。
    *   この`Exporter`は、ページネーション処理、ユーザー名解決（キャッシュ付き）、ファイルダウンロード処理、データ形式の変換などのロジックを含みます。

#### フェーズ3：コマンドラインインターフェース（CLI）統合

1.  **タイムスタンプ処理の仕様:**
    *   **タイムゾーン指定なしの場合:** ユーザーが`--start-time`や`--end-time`でタイムゾーンオフセットを指定しない場合（例: `2025-08-12T10:00:00`）、コマンドが実行されたシステムのローカルタイムとして解釈します。
    *   **タイムゾーン指定ありの場合:** オフセット（`+09:00`）やZulu（`Z`）が指定された場合は、それを正確な時刻として解釈します。

2.  **新しいCobraコマンドの作成:**
    *   **ファイル:** `cmd/export.go` (親コマンド) および `cmd/export_log.go` (サブコマンド)。
    *   **コマンド構造:** `scat export log`
    *   **フラグ:**
        *   `--channel <name>` (必須): エクスポート元のチャネル。
        *   `--output-format <format>` (任意, デフォルト `json`): `json` または `text`。
        *   `--start-time <timestamp>` (任意): 時間範囲の開始。フォーマット: RFC3339。
        *   `--end-time <timestamp>` (任意): 時間範囲の終了。フォーマット: RFC3339。
        *   `--include-files` (任意, boolean): 添付ファイルをダウンロードするかどうか。
        *   `--output-dir <path>` (任意): エクスポートされたファイルを保存するディレクトリ。**デフォルト: `./scat-export-<UTCタイムスタンプ>/`**。

3.  **コマンドロジックの実装:**
    *   フラグを解析し、タイムスタンプを解釈します。
    *   `--silent`モードでない限り、解釈されたエクスポート期間を示すステータスメッセージを**一行で**表示します。
        *   **メッセージ例:** `> Exporting messages from 2025-08-13T10:00:00+09:00 to 2025-08-14T10:00:00+09:00 (UTC: 2025-08-13T01:00:00Z to 2025-08-14T01:00:00Z)`
    *   `provider.Capabilities().CanExportLogs`をチェックし、サポートされていなければエラー終了します。
    *   `provider.LogExporter()`を呼び出してエクスポーターを取得し、`internal/export.Exporter`をインスタンス化します。
    *   `Exporter`のメインメソッドを呼び出し、結果をファイルにストリーミング出力します。

#### フェーズ4：ドキュメントと検証

1.  **`README.md`と`README.ja.md`の更新:**
    *   `export log`コマンド、その使用法、およびフラグの新しいセクションを追加します。
2.  **`SLACK_SETUP.md`の更新:**
    *   エクスポートに必要なSlack APIスコープを追加します: `conversations:history`、`users:read`、`files:read`。
3.  **`make lint`と`make test`の実行:**
    *   すべての新しいコードがリンティングルールに準拠し、すべてのテストがパスすることを確認します。
4.  **手動テスト:**
    *   さまざまなオプションを使用して`scat export log`コマンドのエンドツーエンドテストを実行します。

### VI. 推定工数

*   **フェーズ1（プロバイダーインターフェースとAPI統合）:** 6〜10時間
*   **フェーズ2（コアロジックとデータ構造）:** 8〜14時間
*   **フェーズ3（CLI統合）:** 4〜8時間
*   **フェーズ4（ドキュメントと検証）:** 2〜4時間

**合計推定工数:** 20〜36時間。

### VII. データ検証とサニタイズ

*   **パス・トラバーサル対策:** ユーザーが提供するすべてのパス（`--output-dir`）は、`filepath.Clean`と`filepath.Join`を使用してクリーンアップされ、絶対パスに解決されます。ダウンロードされたファイル名はサニタイズされます。
*   **構造化データサニタイズ:** 最終的なJSON出力に埋め込まれるすべての文字列コンテンツは、適切なUTF-8であることが検証されます。URLは正しいフォーマットであることが検証されます。

### VIII. 未解決の質問/考慮事項 (最終版)

*   **レート制限:** Slack APIにはレート制限があります。**`429 Too Many Requests`エラーを適切に処理するために、ベースAPIクライアントに指数バックオフによるリトライメカニズムを実装する必要があります。**
*   **エラー処理:** API呼び出し、ファイル操作、データ解析のための堅牢なエラー処理。
*   **大規模なエクスポート:** 過剰なメモリ使用なしに大規模なエクスポートを効率的に処理するため、**JSON出力はメモリに保持せず、`json.Encoder`を使用して直接ファイルにストリーミングします。** ファイルのダウンロードは順次処理されます。
*   **ユーザー権限:** 必要なSlackボット権限（`conversations:history`、`users:read`、`files:read`）は、`SLACK_SETUP.md`に明確に文書化されます。
*   **タイムゾーン処理:** `--start-time`および`--end-time`フラグの解釈ロジックはフェーズ3で定義済み。
*   **出力ファイル命名:** 明確で一貫した命名規則を使用します（例: `export-<channel>-<タイムスタンプ>.json`）。

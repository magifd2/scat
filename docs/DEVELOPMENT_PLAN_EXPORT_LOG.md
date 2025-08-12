## Development Plan: Channel Log Export Feature

### I. Feature Overview

This feature will allow users to export chat logs from a specified Slack channel. The exported data will be structured (JSON format), include both user IDs and resolved user names, and optionally download attached files to a local directory.

### II. Core Principles Adherence

*   **Security First:** All Slack API tokens will be handled securely. File downloads will be validated. **Data validation and sanitization will be rigorously applied to prevent path traversal, injection, and other vulnerabilities.**
*   **Testability First:** New functions and components will be designed with unit testing in mind.
*   **Explicit is Better than Implicit:** Dependencies will be explicitly passed.
*   **Code Style and Quality:** Adhere to standard Go formatting and idiomatic Go practices.
*   **Dependency Management:** Use Go Modules; `go mod tidy` will be run after adding new dependencies.

### III. Definition of "Major Refactoring"

This feature introduces new functionality and does not involve changes to initialization logic, core interfaces (like `Provider`), or widespread impact across three or more existing packages. Therefore, it is **not** considered a "Major Refactoring" as per `GEMINI.md` III-1.

### IV. Safe Refactoring Protocol

The development will follow the Safe Refactoring Protocol (III-2) by making minimal changes, testing immediately, and committing on success.

### V. Detailed Plan

#### Phase 1: Provider Interface and Slack API Integration

1.  **`Capabilities`構造体の更新:**
    *   **ファイル:** `internal/provider/provider.go`
    *   **変更点:** `Capabilities`構造体に新しいブール値フィールドを追加します。
        *   `CanExportLogs bool`
        *   `CanResolveUsers bool`
        *   `CanDownloadFiles bool`
    *   **テスト:** `Provider`実装の既存のテストが引き続きパスすることを確認します。
2.  **`Provider`インターフェースの更新（新しいメソッドの追加）:**
    *   **ファイル:** `internal/provider/provider.go`
    *   **変更点:** `Provider`インターフェースに新しいメソッドを追加します。
        *   `GetConversationHistory(channelID string, latest, oldest string, limit int, cursor string) (*ConversationsHistoryResponse, error)`
        *   `GetUserInfo(userID string) (*UserInfoResponse, error)`
        *   `DownloadFile(fileURL string, token string) ([]byte, error)`
    *   **テスト:** `Provider`実装の既存のテストが引き続きパスすることを確認します。
3.  **Slackプロバイダーでの新しいメソッドと機能の実装:**
    *   **ファイル:** `internal/provider/slack/api.go`（API呼び出し用）および`internal/provider/slack/slack.go`（`Capabilities`メソッド用）
    *   **関数:** `Provider`インターフェースで定義されている`GetConversationHistory`、`GetUserInfo`、`DownloadFile`メソッドを実装します。
    *   **機能:** `internal/provider/slack/slack.go`の`Capabilities()`メソッドを更新し、`CanExportLogs`、`CanResolveUsers`、`CanDownloadFiles`に対して`true`を返します。
    *   **詳細:** `conversations.history`、`users.info`へのAPI呼び出しとファイルダウンロードロジックを実装します。`conversations.history`のページネーションを処理します。
    *   **テスト:** これらの新しいAPI呼び出しとファイルダウンロードの単体テスト。
4.  **モックプロバイダーの更新:**
    *   **ファイル:** `internal/provider/mock/mock.go`
    *   **変更点:** 新しい`GetConversationHistory`、`GetUserInfo`、`DownloadFile`メソッドを実装します。これらはテスト目的でno-opまたはダミーデータを返すことができます。
    *   **機能:** `Capabilities()`メソッドを更新し、`CanExportLogs`、`CanResolveUsers`、`CanDownloadFiles`に対して適切なブール値を返します。モックプロバイダーの場合、CLIロジックのテストを許可するために`true`を返すことも、サポートされていない機能のCLIの処理をテストするために`false`を返すこともできます。テスト目的で`true`から始めます。
    *   **テスト:** モックプロバイダーの既存のテストが引き続きパスすることを確認します。

#### Phase 2: Core Logic and Data Structs

1.  **Define Structured Data Models:**
    *   **ファイル:** `internal/provider/slack/types.go`（Slack API応答構造体用）および`internal/export/types.go`（汎用エクスポートデータ構造体用の新規ファイル）
    *   **構造体:**
        *   `ExportedLog`: `ChannelInfo`、`ExportTimestamp`、`Messages`を含むトップレベルの構造体。
        *   `ExportedMessage`: `ID`、`UserID`、`UserName`、`Text`、`Timestamp`、`Type`、`Files`（`ExportedFile`の配列）を含む単一のメッセージを表す。
        *   `ExportedFile`: `ID`、`Name`、`MimeType`、`LocalPath`（ダウンロードされた場合）を含む添付ファイルを表す。
    *   **詳細:** 解析および変換されたデータを保持するためのGo構造体を設計します。
2.  **キャッシュ付きユーザーリゾルバーの実装:**
    *   **ファイル:** `internal/export/userresolver.go`（新規ファイル）
    *   **関数:** `ResolveUserName(userID string, provider provider.Interface) (string, error)`
    *   **詳細:** 冗長なAPI呼び出しを避けるために、ユーザーIDのキャッシュメカニズムを実装します。これは`provider.GetUserInfo`メソッドを使用します。
    *   **テスト:** ユーザーリゾルバーとキャッシュの単体テスト。
3.  **ファイルハンドラーの実装:**
    *   **ファイル:** `internal/export/filehandler.go`（新規ファイル）
    *   **関数:** `HandleAttachedFiles(files []SlackFile, exportDir string, provider provider.Interface) ([]ExportedFile, error)`
    *   **詳細:** 添付ファイルを反復処理し、`exportDir`に`provider.DownloadFile`を使用してダウンロードし、`ExportedFile`を`LocalPath`で更新します。

#### Phase 3: Command-Line Interface (CLI) Integration

1.  **Create New Cobra Commands:**
    *   **ファイル:** `cmd/export.go`（親`export`コマンド用の新規ファイル）
    *   **ファイル:** `cmd/export_log.go`（`log`サブコマンド用の新規ファイル）
    *   **コマンド構造:** `scat export log <channel-name>`
    *   **フラグ（`log`サブコマンド用）:**
        *   `--channel <name>`（必須）：エクスポート元のチャネル。
        *   `--output-format <format>`（オプション、デフォルト`json`）：`json`または`text`。
        *   `--start-time <timestamp>`（オプション）：時間範囲の開始。
        *   `--end-time <timestamp>`（オプション）：時間範囲の終了。
        *   `--include-files`（オプション、ブール値）：添付ファイルをダウンロードするかどうか。
        *   `--output-dir <path>`（オプション）：エクスポートされたファイルとJSON出力を保存するディレクトリ。
    *   **詳細:** `exportCmd`を`rootCmd`と統合し、`exportLogCmd`を`exportCmd`と統合します。
2.  **Implement Command Logic:**
    *   Parse flags.
    *   現在のプロバイダーインスタンスを取得します。
    *   **`provider.Capabilities()`を使用してプロバイダーの機能を確認します。**
        *   `provider.Capabilities().CanExportLogs`が`false`の場合、エラーを出力して終了します。
        *   `--include-files`が設定されており、`provider.Capabilities().CanDownloadFiles`が`false`の場合、エラーを出力して終了します。
        *   ユーザー解決が必要で、`provider.Capabilities().CanResolveUsers`が`false`の場合、適切に処理します（例: 名前ではなくユーザーIDを出力するか、警告を出力します）。
    *   チャネル名をIDに解決します。
    *   `Provider.GetConversationHistory`を呼び出します。
    *   メッセージを反復処理し、`UserResolver`を使用してユーザー名を解決し、`FileHandler`を使用してファイルのダウンロードを処理します。
    *   **データ検証とサニタイズを適用します。**
    *   `--output-format`に基づいてデータをフォーマットして出力します。

#### Phase 4: Documentation and Verification

1.  **Update `README.md` and `README.ja.md`:**
    *   Add a new section for the `export` command and its `log` subcommand, their usage, and flags.
2.  **Run `make lint` and `make test`:**
    *   Ensure all new code adheres to linting rules and all tests pass.
3.  **Manual Testing:**
    *   Perform end-to-end testing of the `scat export log` command with various options.

### VI. Estimated Effort

*   **Phase 1 (Provider Interface & Slack API Integration):** 6〜10時間
*   **Phase 2 (Core Logic & Data Structs):** 8〜16時間
*   **Phase 3 (CLI Integration):** 4〜8時間
*   **Phase 4 (Documentation & Verification):** 2〜4時間

**Total Estimated Effort:** 20〜38時間（約2.5〜5日間の集中的な作業）。

### VII. Data Validation and Sanitization (New Section)

To ensure security and data integrity, the following validation and sanitization measures will be implemented:

*   **Path Traversal Prevention:**
    *   When creating local files (e.g., for downloaded attachments or the JSON output file), all user-provided paths (e.g., `--output-dir`) will be sanitized and resolved to absolute paths.
    *   The `filepath.Clean` and `filepath.Join` functions from Go's standard library will be used to prevent malicious path manipulation (e.g., `../../`).
    *   Downloaded filenames will be sanitized to remove any characters that could lead to path traversal or invalid filenames on various operating systems.
*   **Structured Data Sanitization (for JSON output):**
    *   While JSON encoding generally handles special characters, it's crucial to ensure that the *content* being embedded into the JSON (e.g., message text, filenames, URLs from Slack) does not contain control characters or sequences that could be misinterpreted by downstream parsers or lead to display issues.
    *   Specifically, for message text and other string fields, we will ensure they are valid UTF-8.
    *   URLs will be validated to ensure they are well-formed before being included in the structured output.
    *   Any sensitive information (e.g., API tokens) will be explicitly excluded from the output.

### VIII. Open Questions / Considerations (Expanded)

*   **Rate Limiting:** Slack API has rate limits. The implementation should consider strategies to handle them (e.g., exponential backoff).
*   **Error Handling:** Robust error handling for API calls, file operations, and data parsing. Specific error types will be defined where appropriate to allow for more granular error handling and user feedback.
*   **Large Exports:** How to handle very large exports (e.g., thousands of messages, many files) efficiently without excessive memory usage.
    *   For JSON output, consider streaming the output to a file instead of holding the entire structure in memory for very large exports.
    *   For file downloads, process them one by one to manage memory and avoid overwhelming the system.
*   **User Permissions:** Clearly document the required Slack bot permissions (`conversations:history`, `users:read`, `files:read`).
*   **Time Zone Handling:** When dealing with `start-time` and `end-time` flags, clarify how time zones will be handled (e.g., UTC, local time, or requiring a specific format with timezone information).
*   **Output File Naming:** Define a clear and consistent naming convention for the exported JSON file and downloaded attachments.
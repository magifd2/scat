## 開発計画: チャネルログエクスポート機能

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

#### フェーズ1：プロバイダーインターフェースとSlack API統合

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

#### フェーズ2：コアロジックとデータ構造

1.  **構造化データモデルの定義:**
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

#### フェーズ3：コマンドラインインターフェース（CLI）統合

1.  **新しいCobraコマンドの作成:**
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
2.  **コマンドロジックの実装:**
    *   フラグを解析します。
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

#### フェーズ4：ドキュメントと検証

1.  **`README.md`と`README.ja.md`の更新:**
    *   `export`コマンドとその`log`サブコマンド、それらの使用法、およびフラグの新しいセクションを追加します。
2.  **`make lint`と`make test`の実行:**
    *   すべての新しいコードがリンティングルールに準拠し、すべてのテストがパスすることを確認します。
3.  **手動テスト:**
    *   さまざまなオプションを使用して`scat export log`コマンドのエンドツーエンドテストを実行します。

### VI. 推定工数

*   **フェーズ1（プロバイダーインターフェースとSlack API統合）：** 6〜10時間
*   **フェーズ2（コアロジックとデータ構造）：** 8〜16時間
*   **フェーズ3（CLI統合）：** 4〜8時間
*   **フェーズ4（ドキュメントと検証）：** 2〜4時間

**合計推定工数：** 20〜38時間（約2.5〜5日間の集中的な作業）。

### VII. データ検証とサニタイズ

セキュリティとデータ整合性を確保するために、以下の検証とサニタイズ対策が実装されます。

*   **パストラバーサル対策:**
    *   ローカルファイル（ダウンロードされた添付ファイルやJSON出力ファイルなど）を作成する際、ユーザーが提供するすべてのパス（例: `--output-dir`）はサニタイズされ、絶対パスに解決されます。
    *   Goの標準ライブラリの`filepath.Clean`および`filepath.Join`関数を使用して、悪意のあるパス操作（例: `../../`）を防ぎます。
    *   ダウンロードされたファイル名は、パストラバーサルやさまざまなオペレーティングシステムでの無効なファイル名につながる可能性のある文字を削除するためにサニタイズされます。
*   **構造化データサニタイズ（JSON出力用）:**
    *   JSONエンコーディングは通常、特殊文字を処理しますが、JSONに埋め込まれるコンテンツ（例: メッセージテキスト、ファイル名、SlackからのURL）に、ダウンストリームパーサーによって誤って解釈されたり、表示の問題につながる可能性のある制御文字やシーケンスが含まれていないことを確認することが重要です。
    *   具体的には、メッセージテキストやその他の文字列フィールドについては、有効なUTF-8であることを確認します。
    *   URLは、構造化出力に含まれる前に、適切に形成されていることを確認するために検証されます。
    *   機密情報（例: APIトークン）は、出力から明示的に除外されます。

### VIII. 未解決の質問/考慮事項

*   **レート制限:** Slack APIにはレート制限があります。実装では、それらを処理するための戦略（例: 指数バックオフ）を考慮する必要があります。
*   **エラー処理:** API呼び出し、ファイル操作、データ解析のための堅牢なエラー処理。よりきめ細かなエラー処理とユーザーフィードバックを可能にするために、適切な場合は特定のタイプのエラーが定義されます。
*   **大規模なエクスポート:** 非常に大規模なエクスポート（例: 数千のメッセージ、多数のファイル）を過剰なメモリ使用なしで効率的に処理する方法。
    *   JSON出力の場合、非常に大規模なエクスポートでは、構造全体をメモリに保持するのではなく、出力をファイルにストリーミングすることを検討します。
    *   ファイルのダウンロードの場合、メモリを管理し、システムに過負荷をかけないように、1つずつ処理します。
*   **ユーザー権限:** 必要なSlackボット権限（`conversations:history`、`users:read`、`files:read`）を明確に文書化します。
*   **タイムゾーン処理:** `--start-time`および`--end-time`フラグを扱う際、タイムゾーンがどのように処理されるか（例: UTC、ローカルタイム、またはタイムゾーン情報を含む特定の形式の要求）を明確にします。
*   **出力ファイル命名:** エクスポートされたJSONファイルとダウンロードされた添付ファイルの明確で一貫した命名規則を定義します。
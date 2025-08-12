## Development Plan: Channel Log Export Feature (Final)

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

#### Phase 1: Provider Interface and API Integration

1.  **Update `Capabilities` Struct:**
    *   **File:** `internal/provider/provider.go`
    *   **Changes:** To maintain consistency with existing code, add a new boolean field to the `Capabilities` struct.
        ```go
        type Capabilities struct {
            // ... existing fields
            CanExportLogs bool
        }
        ```

2.  **Define New `LogExporter` Interface:**
    *   **File:** `internal/provider/provider.go`
    *   **Changes:** To separate concerns, a new interface dedicated to export functionality will be created.
        ```go
        type LogExporter interface {
            GetConversationHistory(channelID string, latest, oldest string, limit int, cursor string) (*ConversationHistoryResponse, error)
            GetUserInfo(userID string) (*UserInfoResponse, error)
            DownloadFile(fileURL string) ([]byte, error)
        }
        ```
    *   The main `provider.Interface` will be updated with a method to retrieve the exporter.
        ```go
        type Interface interface {
            Capabilities() Capabilities
            // ... existing methods
            LogExporter() LogExporter // To be called only if CanExportLogs is true
        }
        ```

3.  **Define Provider-Level API Data Structures:**
    *   **File:** `internal/provider/types.go` (New File)
    *   **Changes:** Create this file to hold common, provider-agnostic API response structures returned by the `LogExporter` interface.
        *   `ConversationHistoryResponse`
        *   `UserInfoResponse`

4.  **Implement in Slack Provider:**
    *   **File:** `internal/provider/slack/slack.go` and `internal/provider/slack/api.go`
    *   **Changes:**
        *   Update the `Capabilities()` method to return `CanExportLogs: true`.
        *   Implement the `LogExporter()` method to return a concrete type that satisfies the `LogExporter` interface.
        *   Implement `GetConversationHistory`, `GetUserInfo`, and `DownloadFile`.

5.  **Update Mock Provider:**
    *   **File:** `internal/provider/mock/mock.go`
    *   **Changes:**
        *   Update `Capabilities()` to return `CanExportLogs: true` (for testing).
        *   Implement `LogExporter()` to return a mock implementation of `LogExporter`.

#### Phase 2: Core Logic and Data Structs

1.  **Define Generic Export Data Models:**
    *   **File:** `internal/export/types.go` (New File)
    *   **Structs:** Define generic structs for the final output format, such as `ExportedLog`, `ExportedMessage`, `ExportedFile`.

2.  **Implement Core `Exporter` Logic:**
    *   **File:** `internal/export/exporter.go` (New File)
    *   **Changes:** Create a central `Exporter` struct that orchestrates the entire export process. It will take a `provider.LogExporter` in its constructor.
    *   This `Exporter` will contain logic for pagination, user name resolution (with caching), file downloads, and data transformation.

#### Phase 3: Command-Line Interface (CLI) Integration

1.  **Timestamp Handling Specification:**
    *   **No Timezone Provided:** If the user does not specify a timezone offset for `--start-time` or `--end-time` (e.g., `2025-08-12T10:00:00`), it will be interpreted as the local time of the system where the command is run.
    *   **Timezone Provided:** If an offset (`+09:00`) or Zulu (`Z`) is specified, it will be interpreted as that exact time.

2.  **Create New Cobra Commands:**
    *   **File:** `cmd/export.go` (parent command) and `cmd/export_log.go` (subcommand).
    *   **Command Structure:** `scat export log`
    *   **Flags:**
        *   `--channel <name>` (required): Channel to export from.
        *   `--output-format <format>` (optional, default `json`): `json` or `text`.
        *   `--start-time <timestamp>` (optional): Start of time range. Format: RFC3339.
        *   `--end-time <timestamp>` (optional): End of time range. Format: RFC3339.
        *   `--include-files` (optional, boolean): Whether to download attached files.
        *   `--output-dir <path>` (optional): Directory to save exported files. **Defaults to `./scat-export-<UTC-timestamp>/`**.

3.  **Implement Command Logic:**
    *   Parse flags and interpret timestamps according to the specification.
    *   Unless in `--silent` mode, print a **single-line** status message indicating the interpreted export period.
        *   **Example Message:** `> Exporting messages from 2025-08-13T10:00:00+09:00 to 2025-08-14T10:00:00+09:00 (UTC: 2025-08-13T01:00:00Z to 2025-08-14T01:00:00Z)`
    *   Check `provider.Capabilities().CanExportLogs` and exit with an error if not supported.
    *   Call `provider.LogExporter()` to get the exporter and instantiate the `internal/export.Exporter`.
    *   Call the main method on the `Exporter` and stream the results to the output file.

#### Phase 4: Documentation and Verification

1.  **Update `README.md` and `README.ja.md`:**
    *   Add a new section for the `export log` command, its usage, and flags.
2.  **Update `SLACK_SETUP.md`:**
    *   Add the required Slack API scopes for exporting: `channels:history`, `groups:history`, `users:read`, and `files:read`.
3.  **Run `make lint` and `make test`:**
    *   Ensure all new code adheres to linting rules and all tests pass.
4.  **Manual Testing:**
    *   Perform end-to-end testing of the `scat export log` command with various options.

### VI. Estimated Effort

*   **Phase 1 (Provider Interface & API Integration):** 6-10 hours
*   **Phase 2 (Core Logic & Data Structs):** 8-14 hours
*   **Phase 3 (CLI Integration):** 4-8 hours
*   **Phase 4 (Documentation & Verification):** 2-4 hours

**Total Estimated Effort:** 20-36 hours.

### VII. Data Validation and Sanitization

*   **Path Traversal Prevention:** All user-provided paths (`--output-dir`) will be cleaned and resolved to absolute paths using `filepath.Clean` and `filepath.Join`. Downloaded filenames will be sanitized.
*   **Structured Data Sanitization:** All string content embedded in the final JSON output will be validated as proper UTF-8. URLs will be validated for correct formatting.

### VIII. Open Questions / Considerations (Final)

*   **Rate Limiting:** The Slack API has rate limits. **The implementation must include an exponential backoff retry mechanism in the base API client** to handle `429 Too Many Requests` errors gracefully.
*   **Error Handling:** Robust error handling for API calls, file operations, and data parsing.
*   **Large Exports:** To handle large exports efficiently without excessive memory usage, **the JSON output will be streamed directly to a file using `json.Encoder`** instead of being held in memory. File downloads will be processed sequentially.
*   **User Permissions:** The required Slack bot permissions (`channels:history`, `groups:history`, `users:read`, `files:read`) will be clearly documented in `SLACK_SETUP.md`.
*   **Time Zone Handling:** The timestamp interpretation logic is defined in Phase 3.
*   **Output File Naming:** A clear and consistent naming convention will be used (e.g., `export-<channel>-<timestamp>.json`).
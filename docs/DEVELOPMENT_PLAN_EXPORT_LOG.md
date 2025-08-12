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
            GetConversationHistory(opts GetConversationHistoryOptions) (*ConversationHistoryResponse, error)
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
        *   `GetConversationHistoryOptions`
        *   `ConversationHistoryResponse`
        *   `UserInfoResponse`

4.  **Implement in Slack Provider:**
    *   **File:** `internal/provider/slack/slack.go` and related files.
    *   **Changes:** Implement the `LogExporter` interface methods. The `GetConversationHistory` method will internally resolve the channel name to an ID.

5.  **Update Mock Provider:**
    *   **File:** `internal/provider/mock/mock.go`
    *   **Changes:** Update the mock provider to satisfy the updated interfaces.

#### Phase 2: Core Logic and Data Structs

1.  **Define Generic Export Data Models:**
    *   **File:** `internal/export/types.go` (New File)
    *   **Structs:** Define generic structs for the final output format, such as `ExportedLog`, `ExportedMessage` (with both `Timestamp` and `TimestampUnix` fields), `ExportedFile`.

2.  **Implement Core `Exporter` Logic:**
    *   **File:** `internal/export/exporter.go` (New File)
    *   **Changes:** Create a central `Exporter` struct that orchestrates the entire export process. It will contain logic for pagination, user name resolution (with caching), file downloads, and data transformation (including timestamp conversion and mention resolution).

#### Phase 3: Command-Line Interface (CLI) Integration

1.  **Timestamp Handling Specification:**
    *   **No Timezone Provided:** If the user does not specify a timezone offset for `--start-time` or `--end-time` (e.g., `2025-08-12T10:00:00`), it will be interpreted as the local time of the system where the command is run.
    *   **Timezone Provided:** If an offset (`+09:00`) or Zulu (`Z`) is specified, it will be interpreted as that exact time.

2.  **Output Handling Specification:**
    *   **Log Output:** A new `--output <path>` flag will specify the destination for the main log. If `<path>` is `-` or the flag is omitted, output is sent to `stdout`.
    *   **File Download Output:** A new `--files-dir <path>` flag will specify the destination for downloaded attachments when `--include-files` is used. If omitted, it defaults to `./scat-export-<channel>-<timestamp>/`.

3.  **Create New Cobra Commands:**
    *   **File:** `cmd/export.go` (parent command) and `cmd/export_log.go` (subcommand).
    *   **Command Structure:** `scat export log`
    *   **Flags:** Update flags to match the new output handling specification (`--output`, `--files-dir`).

4.  **Implement Command Logic:**
    *   Parse flags and interpret timestamps.
    *   Unless in `--silent` mode, print a single-line status message indicating the export parameters.
    *   Handle the logic for directing the main log output to either `stdout` or a file based on the `--output` flag.
    *   Instantiate and run the `Exporter`.
    *   Write the returned data to the specified output.

#### Phase 4: Documentation and Verification

1.  **Update `README.md` and `README.ja.md`:**
    *   Add a new section for the `export log` command, its usage, and flags, including clear examples for the new output redirection features.
2.  **Update `SLACK_SETUP.md`:**
    *   Add the required Slack API scopes for exporting: `channels:history`, `groups:history`, `users:read`, and `files:read`.
3.  **Run `make lint` and `make test`:**
    *   Ensure all new code adheres to linting rules and all tests pass.
4.  **Manual Testing:**
    *   Perform end-to-end testing of the `scat export log` command with various options.

### VI. Estimated Effort

*   **Phase 1 (Provider Interface & API Integration):** 6-10 hours
*   **Phase 2 (Core Logic & Data Structs):** 8-14 hours
*   **Phase 3 (CLI Integration):** 5-9 hours (slightly increased for new flag logic)
*   **Phase 4 (Documentation & Verification):** 3-5 hours (slightly increased for new examples)

**Total Estimated Effort:** 22-38 hours.

### VII. Data Validation and Sanitization

*   **Path Traversal Prevention:** All user-provided paths will be cleaned and resolved to absolute paths. Downloaded filenames will be sanitized.
*   **Structured Data Sanitization:** All string content embedded in the final JSON output will be validated as proper UTF-8. URLs will be validated.

### VIII. Open Questions / Considerations (Final)

*   **Rate Limiting:** The implementation must include an exponential backoff retry mechanism.
*   **Error Handling:** Robust error handling for all operations.
*   **Large Exports:** JSON output will be streamed. File downloads will be sequential.
*   **User Permissions:** Required permissions (`channels:history`, `groups:history`, `users:read`, `files:read`) will be documented.
*   **Timestamp Handling:** Logic is defined in Phase 3. The final JSON will contain both human-readable and Unix timestamps.
*   **Output File Naming:** A clear and consistent naming convention will be used.

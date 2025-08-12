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

1.  **Update `Capabilities` Struct:**
    *   **File:** `internal/provider/provider.go`
    *   **Changes:** Add new boolean fields to the `Capabilities` struct:
        *   `CanExportLogs bool`
        *   `CanResolveUsers bool`
        *   `CanDownloadFiles bool`
    *   **Testing:** Ensure existing tests for `Provider` implementations still pass.
2.  **Update `Provider` Interface (Add New Methods):**
    *   **File:** `internal/provider/provider.go`
    *   **Changes:** Add new methods to the `Provider` interface:
        *   `GetConversationHistory(channelID string, latest, oldest string, limit int, cursor string) (*ConversationsHistoryResponse, error)`
        *   `GetUserInfo(userID string) (*UserInfoResponse, error)`
        *   `DownloadFile(fileURL string, token string) ([]byte, error)`
    *   **Testing:** Ensure existing tests for `Provider` implementations still pass.
3.  **Implement New Methods and Capabilities in Slack Provider:**
    *   **File:** `internal/provider/slack/api.go` (for API calls) and `internal/provider/slack/slack.go` (for `Capabilities` method)
    *   **Functions:** Implement the `GetConversationHistory`, `GetUserInfo`, and `DownloadFile` methods as defined in the `Provider` interface.
    *   **Capabilities:** Update the `Capabilities()` method in `internal/provider/slack/slack.go` to return `true` for `CanExportLogs`, `CanResolveUsers`, and `CanDownloadFiles`.
    *   **Details:** Implement the API calls to `conversations.history`, `users.info`, and the file download logic. Handle pagination for `conversations.history`.
    *   **Testing:** Unit tests for these new API calls and file download.
4.  **Update Mock Provider:**
    *   **File:** `internal/provider/mock/mock.go`
    *   **Changes:** Implement the new `GetConversationHistory`, `GetUserInfo`, and `DownloadFile` methods. These can be no-op or return dummy data for testing purposes.
    *   **Capabilities:** Update the `Capabilities()` method to return appropriate boolean values for `CanExportLogs`, `CanResolveUsers`, and `CanDownloadFiles`. For the mock provider, these might return `true` to allow testing the CLI logic, or `false` to test the CLI's handling of unsupported features. We'll start with `true` for testing purposes.
    *   **Testing:** Ensure existing tests for the mock provider still pass.

#### Phase 2: Core Logic and Data Structs

1.  **Define Structured Data Models:**
    *   **File:** `internal/provider/slack/types.go` (for Slack API response structs) and `internal/export/types.go` (new file for generic export data structs)
    *   **Structs:**
        *   `ExportedLog`: Top-level struct containing `ChannelInfo`, `ExportTimestamp`, `Messages`.
        *   `ExportedMessage`: Represents a single message, including `ID`, `UserID`, `UserName`, `Text`, `Timestamp`, `Type`, `Files` (array of `ExportedFile`).
        *   `ExportedFile`: Represents an attached file, including `ID`, `Name`, `MimeType`, `LocalPath` (if downloaded).
    *   **Details:** Design the Go structs to hold the parsed and transformed data.
2.  **Implement User Resolver with Caching:**
    *   **File:** `internal/export/userresolver.go` (new file)
    *   **Function:** `ResolveUserName(userID string, provider provider.Interface) (string, error)`
    *   **Details:** Implement a caching mechanism for user IDs to avoid redundant API calls. This will use the `provider.GetUserInfo` method.
    *   **Testing:** Unit tests for the user resolver and caching.
3.  **Implement File Handler:**
    *   **File:** `internal/export/filehandler.go` (new file)
    *   **Function:** `HandleAttachedFiles(files []SlackFile, exportDir string, provider provider.Interface) ([]ExportedFile, error)`
    *   **Details:** Iterate through attached files, download them to `exportDir` using `provider.DownloadFile`, and update `ExportedFile` with `LocalPath`.

#### Phase 3: Command-Line Interface (CLI) Integration

1.  **Create New Cobra Commands:**
    *   **File:** `cmd/export.go` (new file for the parent `export` command)
    *   **File:** `cmd/export_log.go` (new file for the `log` subcommand)
    *   **Command Structure:** `scat export log <channel-name>`
    *   **Flags (for `log` subcommand):**
        *   `--channel <name>` (required): Channel to export from.
        *   `--output-format <format>` (optional, default `json`): `json` or `text`.
        *   `--start-time <timestamp>` (optional): Start of time range.
        *   `--end-time <timestamp>` (optional): End of time range.
        *   `--include-files` (optional, boolean): Whether to download attached files.
        *   `--output-dir <path>` (optional): Directory to save exported files and JSON output.
    *   **Details:** Integrate `exportCmd` with `rootCmd`, and `exportLogCmd` with `exportCmd`.
2.  **Implement Command Logic:**
    *   Parse flags.
    *   Get the current provider instance.
    *   **Check provider capabilities using `provider.Capabilities()`:**
        *   If `provider.Capabilities().CanExportLogs` is `false`, print an error and exit.
        *   If `--include-files` is set and `provider.Capabilities().CanDownloadFiles` is `false`, print an error and exit.
        *   If user resolution is required and `provider.Capabilities().CanResolveUsers` is `false`, handle appropriately (e.g., output user IDs instead of names, or print a warning).
    *   Resolve channel name to ID.
    *   Call `Provider.GetConversationHistory`.
    *   Iterate through messages, resolve user names using the `UserResolver`, and handle file downloads using the `FileHandler`.
    *   **Apply data validation and sanitization.**
    *   Format and output the data based on `--output-format`.

#### Phase 4: Documentation and Verification

1.  **Update `README.md` and `README.ja.md`:**
    *   Add a new section for the `export` command and its `log` subcommand, their usage, and flags.
2.  **Run `make lint` and `make test`:**
    *   Ensure all new code adheres to linting rules and all tests pass.
3.  **Manual Testing:**
    *   Perform end-to-end testing of the `scat export log` command with various options.

### VI. Estimated Effort

*   **Phase 1 (Provider Interface & Slack API Integration):** 6-10 hours
*   **Phase 2 (Core Logic & Data Structs):** 8-16 hours
*   **Phase 3 (CLI Integration):** 4-8 hours
*   **Phase 4 (Documentation & Verification):** 2-4 hours

**Total Estimated Effort:** 20-38 hours (approx. 2.5-5 days of focused work).

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
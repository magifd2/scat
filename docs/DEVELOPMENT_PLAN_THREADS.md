# Development Plan: Thread Export for `export log`

This document outlines the detailed design and development plan for adding thread export functionality to the `export log` command.

## 1. Detailed Design

### 1.1. Data Structure Extension (`internal/export/types.go`)

The exported message data structure will be extended to include thread-specific information.

- The following fields will be added to the `export.ExportedMessage` struct:
  - `ThreadTimestampUnix string`: If the message is a reply, this field will store the Unix timestamp of the parent message of the thread. It will be an empty string for parent messages themselves.
  - `IsReply bool`: A flag to indicate whether the message is a reply within a thread.

```go
// internal/export/types.go
type ExportedMessage struct {
	UserID        string         `json:"user_id"`
	UserName      string         `json:"user_name"`
	PostType      string         `json:"post_type"`
	Timestamp     string         `json:"timestamp_rfc3339"`
	TimestampUnix string         `json:"timestamp_unix"`
	Text          string         `json:"text"`
	Files         []ExportedFile `json:"files"`
	// --- Additions ---
	ThreadTimestampUnix string `json:"thread_timestamp_unix,omitempty"`
	IsReply             bool   `json:"is_reply"`
}
```

### 1.2. Slack Provider Implementation (`internal/provider/slack/export.go`)

The implementation will utilize the `conversations.replies` API in addition to the `conversations.history` API to fetch replies within threads.

1.  **Modification of `ExportLog` function**:
    - Call `conversations.history` to fetch the top-level messages in a channel.
    - For each message retrieved, check the `reply_count` field.
    - If `reply_count > 0`, the message is a parent of a thread.
    - Use the message's timestamp (`ts`) to call the `conversations.replies` API and retrieve all messages in that thread (the parent itself and all replies).
    - Map the messages retrieved from `conversations.replies` to the extended `ExportedMessage` struct.
        - For replies, set `IsReply: true` and `ThreadTimestampUnix: (parent's timestamp)`.
    - Manage processed thread parent messages to avoid duplicating them in the final result.

### 1.3. Output Format Adjustment (`cmd/export_log.go`)

The output will be adjusted to make the thread structure more understandable in the exported log file.

-   **JSON Format**: No significant changes are required as the modifications to the `ExportedMessage` struct will be reflected automatically.
-   **Text Format**:
    -   The `saveExportedLog` function will be modified.
    -   First, all messages will be sorted by their timestamp.
    -   When outputting messages, if `IsReply` is `true`, the text will be indented (e.g., with a `	` or 4 spaces) to visually indicate that it is a reply.

---

## 2. Development Plan

This enhancement will be developed in the following phases:

-   **Phase 1: Branch Creation and Plan Documentation (This task)**
    1.  Create the `feature/export-log-threads` branch.
    2.  Save this design and development plan as `docs/DEVELOPMENT_PLAN_THREADS.md` and commit it.

-   **Phase 2: Data Structure and Provider Implementation**
    1.  Modify the `ExportedMessage` struct in `internal/export/types.go` as designed.
    2.  Modify `internal/provider/slack/export.go` to implement the logic for fetching thread replies using `conversations.replies`.
    3.  Run `make test` at this stage to ensure that existing tests are not broken.

-   **Phase 3: Output Processing and Testing**
    1.  Modify `saveExportedLog` in `cmd/export_log.go` to implement indented display for the text format.
    2.  Prepare mock data that includes threads and add new test cases to `internal/provider/slack/slack_provider_test.go` and `cmd/export_log_test.go` to verify that threads are correctly fetched and displayed.

-   **Phase 4: Documentation Update and Final Verification**
    1.  Update `README.md` and `EXPORT_FORMAT.md` to include information about the thread export feature.
    2.  Run `make all` to ensure that the build, tests, and linting all pass.
    3.  Conduct manual testing against a live Slack workspace to verify functionality.

-   **Phase 5: Pull Request Creation**
    1.  Commit all changes and create a pull request to the `main` branch.

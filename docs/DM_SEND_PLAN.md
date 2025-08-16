# DM Send Feature Development Plan

This document outlines the development plan for implementing the Direct Message (DM) sending feature.

## 1. Investigation Summary

To send a Direct Message (DM) to a specific user via the Slack API, a two-step process is required:

1.  **Obtain DM Channel ID:** Use the `conversations.open` API endpoint with the target user's ID. This returns a unique DM Channel ID (e.g., `D0123ABC`).
2.  **Post Message:** Use the standard `chat.postMessage` API with the obtained DM Channel ID to send the message.

This process requires the `im:write` scope to be added to the Slack App's permissions.

## 2. Development Strategy

To integrate this functionality with minimal changes to the existing codebase, the following approach will be taken:

-   **New `--user` Flag:**
    A `--user <user_id>` flag will be added to the `scat post` command. This flag will be mutually exclusive with the existing `--channel` flag.

-   **Leverage Existing Provider:**
    The user ID will be passed down to the existing `PostMessage` flow. The Slack provider will handle the two-step API logic internally, thus minimizing changes to the core interfaces.

## 3. Development Phases

### Phase 1: Extend Type Definitions
-   **File:** `internal/provider/types.go`
-   **Action:** Add a `UserID string` field to the `PostMessageOptions` struct to hold the target user ID.

### Phase 2: Extend Command-Line Interface
-   **File:** `cmd/post.go`
-   **Actions:**
    -   Define the new `--user` flag.
    -   Implement mutual exclusion logic to prevent `--user` and `--channel` from being used together.
    -   Set the `UserID` field in `PostMessageOptions` with the value from the `--user` flag.

### Phase 3: Implement Slack Provider Logic
-   **File:** `internal/provider/slack/api.go`
-   **Action:** Add a new function to call the `conversations.open` API.
-   **File:** `internal/provider/slack/slack_provider.go`
-   **Action:** Modify the `PostMessage` method. If the `UserID` field is set, call `conversations.open` to get the DM channel ID, then proceed with the `chat.postMessage` call.
-   **Files:** `internal/provider/mock/mock_provider.go`, `internal/provider/test_provider/test_provider.go`
-   **Action:** Update providers to correctly handle (e.g., log) the new `UserID` field in `PostMessageOptions`.

### Phase 4: Add Tests
-   **File:** `cmd/post_test.go`
-   **Actions:**
    -   Add a test case to verify that a DM is sent correctly using the `--user` flag.
    -   Add a test case to ensure the mutual exclusion logic for `--user` and `--channel` works as expected.

### Phase 5: Update Documentation
-   **Files:** `README.md`, `docs/SLACK_SETUP.md`
-   **Actions:**
    -   Document the new `--user` flag.
    -   Add instructions to include the new `im:write` scope in the Slack App configuration.

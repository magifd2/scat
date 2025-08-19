# Development Plan: Channel Management Features

This document outlines the development plan for implementing new channel creation and user invitation features in `scat`.

## 1. Features

- **`scat channel create`**: A new command to create a public or private Slack channel.
  - It will support setting the description and topic.
  - It will support inviting users and user groups upon creation.
- **`scat channel invite`**: A new command to invite users and user groups to an existing channel.
- User specification will support both User IDs (e.g., `U12345`) and mention names (e.g., `@username`).

## 2. Design Policy

- All parameters for provider methods, including required ones, will be passed via a single `Options` struct to maintain consistency with existing patterns (`Post`, `ExportLog`).
- The implementation will be layered, separating low-level API calls from high-level provider logic and CLI command logic.

## 3. Development Steps

### Step 1: Add Capabilities Definitions
- **File:** `internal/provider/provider.go`
- **Action:** Add `CanCreateChannel` and `CanInviteUsers` boolean flags to the `Capabilities` struct.

### Step 2: Define Interfaces and Types
- **File:** `internal/provider/types.go`
- **Action:** Define `CreateChannelOptions` and `InviteToChannelOptions` structs.
- **File:** `internal/provider/provider.go`
- **Action:** Add the following methods to the `provider.Interface`:
  - `CreateChannel(opts CreateChannelOptions) (Channel, error)`
  - `InviteToChannel(channelID string, opts InviteToChannelOptions) error`

### Step 3: Update All Provider Implementations
- **`internal/provider/slack/`**:
  - **`api.go`**: Add new low-level functions to call the `conversations.create` and `conversations.invite` Slack APIs.
  - **`types.go`**: Add structs to unmarshal the responses from the new APIs.
  - **`slack_provider.go`**: Update the `Capabilities()` method to return `true` for the new features.
  - **`channel.go` (or similar)**: Implement the `CreateChannel` and `InviteToChannel` methods, calling the new functions in `api.go`.
- **`internal/provider/mock/mock_provider.go`**: Add dummy implementations for the new interface methods.
- **`internal/provider/testprovider/test_provider.go`**: Add stateful, in-memory fake implementations for the new interface methods.

### Step 4: Implement CLI Commands
- **Files:** Create `cmd/channel_create.go` and `cmd/channel_invite.go`.
- **File:** `cmd/channel.go`
- **Action:**
  - Register the new commands.
  - The command logic will:
    1. Check the provider's capabilities first.
    2. Resolve user mention names to IDs using the existing `ListUsers` functionality.
    3. Build the appropriate `Options` struct from command-line arguments and flags.
    4. Call the corresponding provider method.

### Step 5: Implement Tests
- **Provider Tests (`internal/provider/slack/*_test.go`):**
  - Use `net/http/httptest` to create a mock HTTP server.
  - Write unit tests to verify that the provider methods send the correct requests to the Slack API and handle responses correctly.
- **CMD Tests (`cmd/*_test.go`):**
  - Use the `test_provider` to test the CLI commands.
  - Verify that command flags are parsed correctly and that the `test_provider`'s in-memory state is updated as expected.

### Step 6: Update Documentation
- **`README.md`, `README.ja.md`**: Add usage instructions for the new commands.
- **`docs/SLACK_SETUP.md`, `docs/SLACK_SETUP.ja.md`**: Add the newly required OAuth scopes to the setup instructions:
  - `channels:manage`
  - `groups:write`
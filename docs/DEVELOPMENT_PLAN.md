# Development Plan

This document outlines the development status and future roadmap for `scat`.

---

## Completed Milestones

### v1.4.0

-   **Configurable Config Path**: Implemented the `--config` global option to specify an alternative path for the configuration file, overriding the default `~/.config/scat/config.json`.

### v1.3.0

-   **Provider Interface Refactoring**: Adopted the "Options Struct" pattern for `PostMessage` and `PostFile` methods to improve consistency and extensibility across the provider interface.

### v1.2.0

-   **Log Export Feature**: Implemented the `scat export log` command.
    -   Supports JSON and plain text output formats.
    -   Allows exporting to stdout (for piping) or a specified file via the `--output` flag.
    -   Supports downloading of attached files via the `--output-files` flag.
    -   Provides time range filtering with `--start-time` and `--end-time`.
    -   Resolves user mentions (`<@USERID>`) into human-readable names (`@username`).
    -   Outputs both human-readable (RFC3339) and Unix timestamps for consistency and compatibility.
-   **Major Refactoring**: Significantly improved the internal architecture for better maintainability.
    -   **Provider Interface**: Adopted the "Options Struct" pattern for all provider methods (`PostMessage`, `PostFile`, `ExportLog`) to improve consistency and extensibility.
    -   **Slack Provider**: Decomposed the monolithic `slack.go` file by separating concerns (posting, uploading, exporting, channel logic) into individual files.
    -   **Security**: Hardened file and directory permissions for all created outputs to `0600` (files) and `0700` (directories).

### Pre-v1.2.0

-   **v0.1.9**: Added comprehensive documentation (`BUILD.md`, `CONTRIBUTING.md`) and significantly updated `README.md` and `README.ja.md` with a command reference and clearer examples.
-   **v0.1.8**: Refactored the Slack provider (`internal/provider/slack`) by splitting the monolithic `slack.go` file into smaller, more manageable files (`api.go`, `channel.go`, `types.go`).
-   **v0.1.7**: Implemented a more robust and explicit provider registration system and introduced the `scat config init` command to separate configuration initialization from loading.
-   **v0.1.6**: Optimized the Slack provider by caching the channel list on initialization to reduce API calls.
-   **v0.1.5**: Centralized the handling of global flags (`--debug`, `--noop`, `--silent`) in the root command to ensure consistent behavior across all subcommands.

---

## Future Roadmap

### 1. Testing Framework (Priority: High)

-   **Current Issue**: The project currently lacks a comprehensive suite of automated tests, making it difficult to verify changes and prevent regressions.
-   **Proposed Solution**: Implement a robust testing framework. For the Slack provider, this would involve using the `httptest` package to create a mock Slack API server. This would allow for testing API interactions, error handling (`not_in_channel`, etc.), and payload construction without making real API calls.

### 2. New Providers (Priority: Low)

-   With the improved provider registration system, adding new providers is now much easier. Potential candidates include:
    -   Discord
    -   Microsoft Teams

### 3. Advanced Features (Priority: Low)

-   **Slack Block Kit Support**: Add support for Slack's "Block Kit" to enable posting richer messages containing images, buttons, and other UI elements. This could be implemented via a new flag, such as `scat post --block-kit '{"blocks": ...}'` or `scat post --block-kit @my-blocks.json`.
-   **Persistent Caching**: Implement a file-based caching mechanism for data like channel lists. This would persist the cache between `scat` command invocations, further reducing API calls for users who run the command frequently.
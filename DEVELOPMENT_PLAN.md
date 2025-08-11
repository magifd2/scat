# Development Plan

This document outlines the development status and future roadmap for `scat`.

---

## Project Status (as of v0.1.9)

The project has undergone significant refactoring and stabilization. The core functionality is robust, and the internal architecture has been improved for better maintainability and extensibility.

-   **Provider Logic**: The provider registration mechanism is now explicit and centralized. The Slack provider's internal logic has been made more efficient by caching channel lists and has been split into smaller, more focused files.
-   **Configuration**: Configuration handling is now more explicit. A `scat config init` command has been introduced to prevent unexpected behavior, and commands now provide clear instructions if a config file is not found.
-   **Error Handling**: Global flags (`--debug`, `--silent`, etc.) are handled centrally, and duplicate error messages have been eliminated.
-   **Documentation**: The `README` files have been significantly expanded with detailed setup instructions, usage examples, and a full command reference. `BUILD.md` and `CONTRIBUTING.md` have also been added.

## Completed Milestones

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

### 2. Provider Enhancements (Priority: Medium)

-   **`PostFile` Error Handling**: The `PostFile` method for the Slack provider currently does not handle `not_in_channel` errors with an automatic join and retry, unlike `PostMessage`. This should be implemented for consistent behavior.
-   **Parameter Structs**: Refactor the `provider.Interface` methods (e.g., `PostMessage`) to accept a parameter struct (e.g., `PostMessage(params PostMessageParams)`). This would create a cleaner, more extensible interface, as provider-specific options would no longer need to be passed as individual arguments through the command layer.

### 3. New Providers (Priority: Low)

-   With the improved provider registration system, adding new providers is now much easier. Potential candidates include:
    -   Discord
    -   Microsoft Teams

### 4. Advanced Features (Priority: Low)

-   **Slack Block Kit Support**: Add support for Slack's "Block Kit" to enable posting richer messages containing images, buttons, and other UI elements. This could be implemented via a new flag, such as `scat post --block-kit '{"blocks": ...}'` or `scat post --block-kit @my-blocks.json`.
-   **Persistent Caching**: Implement a file-based caching mechanism for data like channel lists. This would persist the cache between `scat` command invocations, further reducing API calls for users who run the command frequently.

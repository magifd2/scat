
# Development Plan

This document outlines the development status and future roadmap for `scat`.

---

## Completed Milestones

### v1.5.0 (In Progress)

-   **Comprehensive Test Suite**: Implemented a robust testing framework, significantly increasing test coverage and project stability.
    -   Adopted a black-box testing approach using a dedicated `testprovider` to verify command logic without actual API calls.
    -   Added unit tests for all major commands: `post`, `upload`, `export log`, `channel list`, and `config init`.
    -   Added unit tests for the provider factory (`GetProvider`) and the `testprovider` itself.
-   **Code Refinements**: Improved code consistency and clarity.
    -   Refactored provider source files to use a consistent `_provider.go` naming convention.

### v1.4.0

-   **Configurable Config Path**: Implemented the `--config` global option to specify an alternative path for the configuration file, overriding the default `~/.config/scat/config.json`.

### v1.3.0

-   **Provider Interface Refactoring**: Adopted the "Options Struct" pattern for `PostMessage` and `PostFile` methods to improve consistency and extensibility across the provider interface.

### v1.2.0

-   **Log Export Feature**: Implemented the `scat export log` command.
-   **Major Refactoring**: Significantly improved the internal architecture for better maintainability.

### Pre-v1.2.0

-   **v0.1.9**: Added comprehensive documentation.
-   **v0.1.8**: Refactored the Slack provider.
-   **v0.1.7**: Implemented a new provider registration system and `config init` command.
-   **v0.1.6**: Implemented channel list caching in the Slack provider.
-   **v0.1.5**: Centralized global flag handling.

---

## Future Roadmap

### 1. Testing Framework (Completed)

-   **Status**: A comprehensive suite of automated tests has been implemented for all major commands and internal logic, ensuring stability and preventing regressions. The framework uses a dedicated `testprovider` for black-box testing of command-line interfaces.

### 2. New Providers (Priority: Low)

-   With the improved provider registration system, adding new providers is now much easier. Potential candidates include:
    -   Discord
    -   Microsoft Teams

### 3. Advanced Features (Priority: Low)

-   **Slack Block Kit Support**: Add support for Slack's "Block Kit" to enable posting richer messages containing images, buttons, and other UI elements. This could be implemented via a new flag, such as `scat post --block-kit '{"blocks": ...}'` or `scat post --block-kit @my-blocks.json`.
-   **Persistent Caching**: Implement a file-based caching mechanism for data like channel lists. This would persist the cache between `scat` command invocations, further reducing API calls for users who run the command frequently.

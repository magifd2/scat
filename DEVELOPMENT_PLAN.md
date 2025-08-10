# Development Plan

## Project Status

Initial development is complete as of `v0.1.1`. The application provides a CLI for posting text messages and uploading files, with a primary focus on supporting Slack. The architecture has been refactored to use a provider-based model to easily support different services in the future.

---

## Future Improvements (TODO)

### 1. Loosely Coupled Refactoring via Parameter Structs

-   **Current Issue**: The `provider.Interface` methods (e.g., `PostMessage`) take arguments like `iconEmoji` that are only used by specific providers (`slack`). This is a form of leaky abstraction, and the command layer is still loosely aware of provider-specific options.
-   **Proposed Improvement**: Refactor the method signatures to accept a parameter struct (e.g., `PostMessage(params PostMessageParams)`). Each provider would then be responsible for interpreting the fields within the struct that are relevant to it. This would create a cleaner interface and improve maintainability and scalability.

### 2. Slack Block Kit Support (Priority: High)

-   **Current Issue**: `scat post` can only send simple text messages.
-   **Proposed Improvement**: Add support for Slack's "Block Kit" to enable posting richer messages containing images, buttons, and other UI elements. This could be implemented via a new flag, such as `scat post --block-kit '{"blocks": ...}'` or `scat post --block-kit @my-blocks.json`, which would send a `blocks` field instead of a `text` field in the JSON payload.
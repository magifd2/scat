# Unit Testing Framework Development Plan

This document outlines a development plan for establishing a comprehensive unit testing framework for the `scat` project. The primary constraint for this plan is to minimize modifications to the existing main codebase.

## 1. Current Testing Status

The `scat` project currently lacks a comprehensive suite of automated unit tests. While `make test` command exists, it primarily serves to check for compilation errors rather than verifying the correctness of the application's logic. This absence of tests makes it challenging to:

-   Verify the correctness of new features or bug fixes.
-   Prevent regressions when refactoring or adding new functionalities.
-   Ensure code quality and maintainability over time.

## 2. Guiding Principles for Testing

To adhere to the constraint of minimizing changes to the main codebase, the following principles will guide the development of the testing framework:

-   **Black-Box Testing**: Where possible, tests will treat components as black boxes, interacting with their public interfaces without relying on internal implementation details.
-   **Interface-Based Testing**: Go's interface system will be leveraged to create mock or stub implementations of dependencies, allowing for isolated testing of units.
-   **Dependency Injection**: Existing dependency injection patterns (e.g., passing `appcontext.Context` or `config.Profile` to functions/methods) will be utilized to inject test doubles where necessary.
-   **Standard Go Testing Library**: The built-in `testing` package will be the primary tool for writing tests, supplemented by other libraries (e.g., `httptest` for mocking HTTP services) as needed.

## 3. Testing Strategy by Component

### 3.1. `cmd` Package (CLI Commands)

-   **Focus**: Verify that commands parse arguments and flags correctly, call the appropriate internal logic, and handle errors gracefully.
-   **Approach**: Use `cobra.Command`'s testing utilities (if available) or manually construct command execution scenarios. Mock out calls to `os.Exit` and `fmt.Fprintf(os.Stderr, ...)` to capture output and exit codes. Inject mock `provider.Interface` implementations to test command interactions with external services without actual network calls.
-   **Minimal Codebase Changes**: Commands are already designed to receive `appcontext.Context` and `config.Profile`, which can contain mock dependencies.

### 3.2. `internal/config` Package

-   **Focus**: Test the loading, saving, and manipulation of configuration profiles.
-   **Approach**: Create temporary files for configuration data. Test `config.Load()`, `config.Save()`, `config.NewDefaultConfig()`, and profile manipulation functions with various valid and invalid JSON structures. Ensure correct handling of default values and backward compatibility.
-   **Minimal Codebase Changes**: The `config` package is already self-contained and operates on file paths, making it highly testable without internal modifications.

### 3.3. `internal/appcontext` Package

-   **Focus**: Test the creation and retrieval of the application context.
-   **Approach**: Simple unit tests to verify that `NewContext` correctly initializes the `Context` struct and that values can be retrieved from a `context.Context` using `appcontext.CtxKey`.
-   **Minimal Codebase Changes**: No changes expected, as this package is purely data-oriented.

### 3.4. `internal/util` Package

-   **Focus**: Test utility functions (e.g., `time.go`).
-   **Approach**: Standard unit tests for input/output correctness and edge cases.
-   **Minimal Codebase Changes**: No changes expected.

### 3.5. `internal/provider` Package (Interface Implementations)

This is a critical area for testing, as it involves interactions with external services.

#### 3.5.1. `internal/provider/mock`

-   **Focus**: Verify that the mock provider correctly simulates the `provider.Interface` behavior for testing purposes.
-   **Approach**: Test that `PostMessage`, `PostFile`, `ListChannels`, and `ExportLog` methods of the mock provider produce the expected output (e.g., to `os.Stderr`) and return the correct dummy data or errors.
-   **Minimal Codebase Changes**: No changes expected.

#### 3.5.2. `internal/provider/slack`

-   **Focus**: Test the Slack provider's interaction with the Slack API, including request construction, response parsing, and error handling.
-   **Approach**: Utilize Go's `net/http/httptest` package to create a local HTTP server that mimics the Slack API. Tests will send requests to this mock server and assert on the outgoing request (headers, body, URL) and the parsing of the mock server's responses. This eliminates the need for actual network calls and real Slack tokens during testing.
    -   **Mocking Slack API**: The `httptest.Server` will be configured to return predefined JSON responses for various Slack API endpoints (e.g., `chat.postMessage`, `files.upload`, `conversations.history`).
    -   **Error Scenarios**: Test how the Slack provider handles API errors (e.g., `not_in_channel`, `invalid_auth`), rate limits, and network issues.
    -   **Data Transformation**: Verify that data is correctly transformed between `scat`'s internal types (`provider.PostMessageOptions`, `provider.Message`) and Slack's API payload formats.
-   **Minimal Codebase Changes**: The Slack provider's methods already accept `appcontext.Context` and `config.Profile`. The `http.Client` used for API calls can be made configurable (e.g., via a field in the `Provider` struct that can be set in tests to use a mock client or `httptest.Server`'s client), which is a minor and common pattern for testability.

## 4. Benefits of a Testing Framework

Implementing this testing framework will provide several key benefits:

-   **Regression Prevention**: Automated tests will catch unintended side effects of new features or refactoring, ensuring existing functionality remains intact.
-   **Improved Code Quality**: Writing testable code often leads to more modular, loosely coupled, and maintainable code.
-   **Faster Development Cycles**: Developers can make changes with confidence, knowing that tests will quickly identify any breakage.
-   **Easier Onboarding**: New contributors can understand the expected behavior of the codebase by examining the tests.
-   **Enhanced Reliability**: The application will be more robust and less prone to bugs in production.

## 5. Next Steps

Upon approval of this plan, the next steps would involve:

1.  Creating the test directory structure (`<package_name>_test.go` files).
2.  Implementing mock HTTP servers for external dependencies (e.g., Slack API).
3.  Writing unit tests for each component as outlined above.
4.  Integrating test execution into the `Makefile` (already present via `make test`).

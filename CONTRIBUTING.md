# Contributing to scat

First off, thank you for considering contributing to `scat`! Whether it's a bug report, a new feature, or an improvement to the documentation, your help is greatly appreciated.

## How to Contribute

There are several ways you can contribute to this project:

-   **Reporting Bugs**: If you find a bug, please open an issue and provide as much detail as possible.
-   **Suggesting Enhancements**: If you have an idea for a new feature or an improvement to an existing one, open an issue to discuss it.
-   **Submitting Pull Requests**: If you want to contribute code, please submit a pull request.

## Reporting Bugs

When opening an issue for a bug, please include the following:

-   **`scat` version**: Run `scat --version`.
-   **Operating System**: e.g., macOS 14.2, Ubuntu 22.04.
-   **What you did**: The exact command and arguments you used.
-   **What you expected to happen**: A clear description of what you thought would happen.
-   **What actually happened**: A description of the error, including any output or logs. Use `scat --debug ...` to get more detailed logs.

## Submitting Pull Requests

1.  **Fork the repository** and create your branch from `main`.
2.  **Make your changes**. Please ensure your code is formatted with `gofmt`.
3.  **Write clear commit messages**. We follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification. This helps in automating changelogs and makes the project history easier to read.
    -   `feat:` for new features.
    -   `fix:` for bug fixes.
    -   `docs:` for documentation changes.
    -   `refactor:` for code changes that neither fix a bug nor add a feature.
    -   `test:` for adding or improving tests.
4.  **Update the documentation** (`README.md`, etc.) if your changes affect it.
5.  **Submit the pull request**. Provide a clear description of the problem and your solution.

## Development Setup

For instructions on how to build the project, run tests, and other development tasks, please see [BUILD.md](./docs/BUILD.md).

### Code Structure Overview

-   `main.go`: The main entry point of the application.
-   `cmd/`: Contains all the command-line interface logic, using the `cobra` library. Each command and subcommand has its own file.
-   `internal/config/`: Handles loading and saving the configuration file.
-   `internal/provider/`: Defines the `provider.Interface` and contains the specific implementations for different services (like `slack` and `mock`). This is the primary area to look at if you want to add support for a new service.
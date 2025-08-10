# Build Instructions

This document describes how to build `scat` from source code.

## Prerequisites

*   [Go](https://go.dev/doc/install) (version 1.16 or later recommended)
*   [Git](https://git-scm.com/)
*   `make` command

## Build Commands

This project uses a `Makefile` for all build-related tasks.

*   **`make build`**
    *   Builds a single binary for your current OS and architecture.

*   **`make cross-compile`**
    *   Builds binaries for multiple platforms (macOS, Linux, Windows).

*   **`make test`**
    *   Runs all tests.

*   **`make lint`**
    *   Runs the `golangci-lint` linter.

*   **`make clean`**
    *   Cleans up all build artifacts.

*   **`make install`**
    *   Installs the binary to `/usr/local/bin` by default. You can change the path with the `PREFIX` variable (e.g., `make install PREFIX=~`).

*   **`make uninstall`**
    *   Removes the installed binary.

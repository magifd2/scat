# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-08-11

This is the first official stable release of `scat`. It includes a wide range of features and improvements implemented since the initial development.

### Features
- **Provider-Based Architecture**: Implemented a flexible provider model, making it easy to add support for new services beyond Slack. (`eb581c6`)
- **Slack Provider**:
    -   Post text messages and upload files. (`55d0516`)
    -   Automatically join public channels when posting if the bot is not already a member. (`deaa521`)
    -   List available channels with `scat channel list`.
    -   Override username and icon emoji when posting messages. (`92bdcd5`)
- **Comprehensive CLI**:
    -   `post` command for sending text messages from arguments, files, or stdin. (`55d0516`)
    -   `upload` command for sending files from a path or stdin. (`55d0516`)
    -   `post --stream` for continuously streaming content, like logs. (`55d0516`)
    -   Full profile management (`add`, `list`, `use`, `set`, `remove`).
    -   `config init` command for explicit and safe configuration initialization. (`5301d87`)
- **Global Flags**:
    -   `--debug` for verbose logging. (`92bdcd5`)
    -   `--noop` for dry runs.
    -   `--silent` to suppress non-error output.

### Fixes
- **Centralized Flag Handling**: Correctly propagated global flags (`--debug`, `--noop`, `--silent`) to all commands. (`8bfe7c7`)
- **Duplicate Error Messages**: Prevented errors from being printed twice by cobra and main. (`0490dc7`)
- **Slack File Upload**:
    -   Migrated from the deprecated `files.upload` API to the modern `files.getUploadURLExternal` and `files.completeUploadExternal` methods. (`59d5faa`)
    -   Reverted an incorrect implementation of file sharing that caused `invalid_blocks` errors. (`8c8ffe1`)
- **Input Limits**: Added configurable limits for file and stdin sizes to prevent excessively large uploads. (`9624a53`)

### Refactoring
- **Provider Registration**: Implemented an explicit provider registry in the `cmd` package, removing direct dependencies from commands to specific provider implementations. (`169279f`)
- **Slack Provider Modularity**: Split the monolithic `slack.go` file into smaller, more focused files (`api.go`, `channel.go`, `types.go`) for improved maintainability. (`08960c4`)
- **Configuration Handling**: Separated the concerns of loading and initializing configuration. (`5301d87`)
- **Application Context**: Introduced an `appcontext` to pass global settings cleanly, reducing coupling. (`d882399`)
- **Channel Caching**: Optimized the Slack provider to cache the channel list on initialization, significantly reducing API calls. (`99c3a66`)

### Documentation
- **Comprehensive READMEs**: `README.md` and `README.ja.md` were completely overhauled with detailed setup instructions, usage examples, and a full command reference. (`d9ce1fe`, `4e2881e`)
- **Added `BUILD.md`**: Provides clear instructions for building the project from source. (`9f3d670`)
- **Added `CONTRIBUTING.md`**: Outlines guidelines for contributing to the project. (`9f3d670`)
- **Added `SLACK_SETUP.md`**: A dedicated guide for configuring the Slack provider. (`3e0e407`)
- **Updated `DEVELOPMENT_PLAN.md`**: The development plan is now up-to-date with completed milestones and a future roadmap. (`54b6d4e`)
- **Removed Build Status Badge**: Removed the non-functional build status badge until a CI/CD pipeline is implemented. (`c600a31`)

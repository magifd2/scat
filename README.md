# scat

`scat` is a command-line interface tool for posting content from files or stdin to a configured HTTP endpoint. It is inspired by `slackcat` but generalized to work with any compatible webhook or API endpoint.

## Features

*   **Flexible Input**: Pass content via command-line arguments, files, or standard input (pipes).
*   **Streaming Support**: Continuously stream data from stdin to an endpoint.
*   **Profile Management**: Save multiple endpoint configurations (URL, token, etc.) as profiles and easily switch between them.
*   **Simple Authentication**: Uses a static Bearer Token for authentication.
*   **Single Binary**: Easy to distribute and use.

## Installation

Use the provided `Makefile` for easy installation.

*   **Default Installation (System-wide):**
    To install `scat` to `/usr/local/bin` (requires `sudo`):
    ```bash
    sudo make install
    ```

*   **User-local Installation:**
    To install `scat` to `~/bin` (recommended for non-root users, ensure `~/bin` is in your `PATH`):
    ```bash
    make install PREFIX=~
    ```

## Quick Start

1.  **Configure a profile:**

    First, set up a destination endpoint. The default profile is created with placeholder values.

    ```bash
    # Set the endpoint for the default profile
    scat profile set endpoint https://your-webhook-url.com/xxxx

    # Set the authentication token for the default profile
    scat profile set token YOUR_SECRET_TOKEN
    ```

2.  **Post a message:**

    ```bash
    # Post a simple string
    echo "Hello, World!" | scat post

    # Post from a file
    scat post /path/to/your/file.txt

    # Stream a log file
    tail -f /var/log/syslog | scat post --stream
    ```

## Command Reference

### `scat post`

Posts content to the configured endpoint.

| Flag      | Shorthand | Description                                      |
| --------- | --------- | ------------------------------------------------ |
| `--channel` | `-c`      | Profile to use (overrides the active default).   |
| `--stream`  | `-s`      | Stream messages from stdin continuously.         |

### `scat profile`

Manages configuration profiles.

| Subcommand | Description                                       |
| ---------- | ------------------------------------------------- |
| `list`     | List all available profiles.                      |
| `use`      | Switch the active profile.                        |
| `add`      | Add a new profile.                                |
| `set`      | Set a value in the current profile. (keys: `endpoint`, `token`, `username`) |
| `remove`   | Remove a profile.                                 |

## Development

To build from source:

```bash
make build
```

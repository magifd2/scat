# scat

`scat` is a command-line interface tool for posting content from files or stdin to a configured HTTP endpoint. It is inspired by `slackcat` but modernized and refactored for extensibility.

## Features

*   **Clear Separation of Concerns**: Send text messages with `scat post` and upload files with `scat upload`.
*   **Flexible Input**: `post` accepts text from arguments, files, or stdin. `upload` accepts files from a path or stdin.
*   **Provider Model**: Supports different providers (`slack`, `mock`) with a clean interface for future expansion.
*   **Streaming Support**: Continuously stream data from stdin to an endpoint using `scat post --stream`.
*   **Profile Management**: Save multiple endpoint configurations (URL, token, etc.) as profiles and easily switch between them.
*   **Secure Token Input**: Interactively prompts for tokens so they don't appear in your shell history.

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

1.  **Configure a profile for Slack:**

    ```bash
    # Add a new profile for your Slack team
    # You will be prompted to securely enter your Bot Token (starts with xoxb-)
    scat profile add my-slack-workspace --provider slack --channel "#general"
    ```

2.  **Send messages and upload files:**

    ```bash
    # Send a simple text message from an argument
    scat post "Hello, World!"

    # Send a multi-line message from a file
    scat post --from-file ./my_message.txt

    # Upload a file with a comment
    scat upload --file ./image.png --comment "Here is the latest screenshot."

    # Stream a log file as messages
    tail -f /var/log/syslog | scat post --stream
    ```

## Command Reference

### Common Flags

These flags are available on both `post` and `upload` commands:

| Flag      | Shorthand | Description                                      |
| --------- | --------- | ------------------------------------------------ |
| `--profile` | `-p`      | Profile to use (overrides the active default).   |
| `--channel` | `-c`      | Override the destination channel for this post.  |
| `--username`| `-u`      | Override the username for this post.             |
| `--noop`    |           | Dry run, do not actually send.                   |
| `--iconemoji`| `-i`     | Icon emoji to use (slack provider only).         |

### `scat post [message text]`

Posts a text message. Input is read from arguments, `--from-file`, or stdin (in that order).

| Flag        | Shorthand | Description                               |
| ----------- | --------- | ----------------------------------------- |
| `--from-file`|           | Read message body from a file.            |
| `--stream`  | `-s`      | Stream messages from stdin continuously.  |
| `--tee`     | `-t`      | Print stdin to screen while posting.      |

### `scat upload`

Uploads a file.

| Flag       | Shorthand | Description                                      |
| ---------- | --------- | ------------------------------------------------ |
| `--file`   | `-f`      | **Required.** Path to the file, or `-` for stdin. |
| `--filename`| `-n`     | Filename for the upload.                         |
| `--filetype`|           | Filetype for syntax highlighting (e.g., `go`).   |
| `--comment`| `-m`      | A comment to post with the file.                 |

### `scat profile`

Manages configuration profiles.

| Subcommand | Description                                       |
| ---------- | ------------------------------------------------- |
| `list`     | List all available profiles.                      |
| `use`      | Switch the active profile.                        |
| `add`      | Add a new profile.                                |
| `set`      | Set a value in the current profile.               |
| `remove`   | Remove a profile.                                 |

### `scat channel`

| Subcommand | Description                                       |
| ---------- | ------------------------------------------------- |
| `list`     | List available channels for `slack` profiles.     |

## Acknowledgements

This project is heavily inspired by and based on the concepts of [bcicen/slackcat](https://github.com/bcicen/slackcat). The core logic for handling file/stdin streaming and posting was re-implemented with reference to the original `slackcat` codebase. `slackcat` is also distributed under the MIT License.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

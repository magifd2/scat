# scat: A General-Purpose Command-Line Content Poster

`scat` is a versatile command-line interface for sending content from files or standard input to a configured destination, such as Slack. It is inspired by `slackcat` but is designed to be more generic and extensible.

[![Build Status](https://github.com/magifd2/scat/actions/workflows/build.yml/badge.svg)](https://github.com/magifd2/scat/actions/workflows/build.yml)

---

## Features

- **Post text messages**: Send content from arguments, files, or stdin.
- **Upload files**: Upload files from a path or stdin.
- **Stream content**: Continuously stream from stdin, posting messages periodically.
- **Profile management**: Configure multiple destinations and switch between them easily.
- **Extensible providers**: Currently supports Slack and a mock provider for testing.

## Installation

Download the latest binary for your system from the [Releases](https://github.com/magifd2/scat/releases) page.

Alternatively, you can build from source:

```bash
make build
```

## Initial Setup

Before you can start posting, you need to create a configuration file.

1.  **Initialize the config file**:

    Run the following command to create a default configuration file at `~/.config/scat/config.json`:

    ```bash
    scat config init
    ```

2.  **Configure a Profile**:

    The default profile uses a mock provider, which is useful for testing. To post to a real service like Slack, you need to add a new profile.

    For detailed instructions on setting up a Slack profile, please see the **[Slack Setup Guide](./SLACK_SETUP.md)**.

    Here is a quick example of how to add a new Slack profile:

    ```bash
    # This will prompt you to enter your Slack Bot Token securely.
    scat profile add my-slack-workspace --provider slack --channel "#general"
    ```

3.  **Set the Active Profile**:

    Tell `scat` to use your new profile by default:

    ```bash
    scat profile use my-slack-workspace
    ```

## Usage Examples

Here are some common ways to use `scat`.

### Posting Text Messages (`post`)

-   **From an argument**:
    `scat post "Hello from the command line!"`

-   **From standard input (pipe)**:
    `echo "This message was piped." | scat post`

-   **Streaming from standard input**:
    `tail -f /var/log/system.log | scat post --stream`

### Uploading Files (`upload`)

-   **Upload a file from a path**:
    `scat upload --file ./report.pdf`

-   **Upload with a comment**:
    `scat upload --file ./screenshot.png -m "Here is the screenshot you requested."`

-   **Upload from standard input**:
    `cat data.csv | scat upload --file - --filename data.csv`

## Command Reference

### Global Flags

| Flag      | Description                                      |
| --------- | ------------------------------------------------ |
| `--profile <name>` | Use a specific profile for the command.          |
| `--debug`   | Enable verbose debug logging.                    |
| `--silent`  | Suppress success messages.                       |
| `--noop`    | Perform a dry run without sending content.       |

### Main Commands

| Command         | Description                                      |
| --------------- | ------------------------------------------------ |
| `scat post`     | Posts a text message.                            |
| `scat upload`   | Uploads a file.                                  |
| `scat profile`  | Manages configuration profiles.                  |
| `scat config`   | Manages the configuration file itself.           |
| `scat channel`  | Lists channels for supported providers.          |

### `post` Command Flags

| Flag          | Shorthand | Description                               |
| ------------- | --------- | ----------------------------------------- |
| `--from-file` |           | Read message body from a file.            |
| `--stream`    | `-s`      | Stream messages from stdin continuously.  |
| `--tee`       | `-t`      | Print stdin to screen while posting.      |
| `--username`  | `-u`      | Override the username for this post.      |
| `--iconemoji` | `-i`      | Icon emoji to use (Slack provider only).  |

### `upload` Command Flags

| Flag        | Shorthand | Description                                      |
| ----------- | --------- | ------------------------------------------------ |
| `--file`    | `-f`      | **Required.** Path to the file, or `-` for stdin. |
| `--filename`| `-n`      | Filename for the upload.                         |
| `--filetype`|           | Filetype for syntax highlighting (e.g., `go`).   |
| `--comment` | `-m`      | A comment to post with the file.                 |

### `profile` Subcommands

| Subcommand | Description                                      |
| ---------- | ------------------------------------------------ |
| `list`     | List all available profiles.                     |
| `use`      | Set the active profile.                          |
| `add`      | Add a new profile.                               |
| `set`      | Set a value in the current profile.              |
| `remove`   | Remove a profile.                                |

### `config` and `channel` Subcommands

| Command             | Description                                      |
| ------------------- | ------------------------------------------------ |
| `config init`       | Creates a new default configuration file.        |
| `channel list`      | Lists available channels for `slack` profiles.   |

---

## Acknowledgements

This project is heavily inspired by and based on the concepts of [bcicen/slackcat](https://github.com/bcicen/slackcat). The core logic for handling file/stdin streaming and posting was re-implemented with reference to the original `slackcat` codebase. `slackcat` is also distributed under the MIT License.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

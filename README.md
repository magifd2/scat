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

    ```bash
    scat post "Hello from the command line!"
    ```

-   **From standard input (pipe)**:

    ```bash
    echo "This message was piped." | scat post
    ```

-   **From a file**:

    ```bash
    scat post --from-file ./message.txt
    ```

-   **Streaming from standard input**:

    This is useful for monitoring logs. `scat` will buffer lines and post them every few seconds.

    ```bash
    tail -f /var/log/system.log | scat post --stream
    ```

### Uploading Files (`upload`)

-   **Upload a file from a path**:

    ```bash
    scat upload --file ./report.pdf
    ```

-   **Upload with a comment**:

    ```bash
    scat upload --file ./screenshot.png -m "Here is the screenshot you requested."
    ```

-   **Upload from standard input**:

    You must provide a filename for the upload when streaming from stdin.

    ```bash
    cat data.csv | scat upload --file - --filename data.csv
    ```

### Profile Management (`profile`)

-   **List all profiles**:

    The active profile is marked with an asterisk (`*`).

    ```bash
    scat profile list
    ```

-   **Switch the active profile**:

    ```bash
    scat profile use another-profile
    ```

-   **Run a command with a specific profile** (without changing the active one):

    ```bash
    scat --profile personal-slack post "This is for my personal workspace."
    ```

-   **Add or modify a profile setting**:

    ```bash
    scat profile set channel "#random"
    ```

### Global Flags

These flags can be used with any command.

-   `--debug`: Enable verbose debug logging.
-   `--silent`: Suppress success messages.
-   `--noop`: Perform a dry run without actually posting or uploading content.
-   `--profile <name>`: Use a specific profile for the command.
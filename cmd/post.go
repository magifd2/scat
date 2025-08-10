package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/magifd2/scat/internal/client"
	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
)

// Holds the timestamp of the current thread for the lifetime of the command execution.
var currentThreadTS string

var postCmd = &cobra.Command{
	Use:   "post [file]",
	Short: "Post content from a file or stdin to a configured endpoint",
	Long:  `Posts content to the destination specified in the active profile. Content can be read from a file or from standard input if no file is specified.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Determine profile
		profileName, _ := cmd.Flags().GetString("channel")
		if profileName == "" {
			profileName = cfg.CurrentProfile
		}
		profile, ok := cfg.Profiles[profileName]
		if !ok {
			return fmt.Errorf("profile '%s' not found", profileName)
		}

		// Get optional flags
		username, _ := cmd.Flags().GetString("username")
		iconEmoji, _ := cmd.Flags().GetString("iconemoji")
		thread, _ := cmd.Flags().GetBool("thread")
		tee, _ := cmd.Flags().GetBool("tee")
		noop, _ := cmd.Flags().GetBool("noop")

		// Create client
		apiClient := client.NewClient(profile, noop)

		// Handle stream
		stream, _ := cmd.Flags().GetBool("stream")
		if stream {
			if len(args) > 0 {
				return fmt.Errorf("cannot use file argument with --stream flag")
			}
			return handleStream(apiClient, profileName, username, iconEmoji, thread, tee)
		}

		// Handle file post or stdin post
		if len(args) > 0 {
			// Post from file (multipart)
			filePath := args[0]
			filename, _ := cmd.Flags().GetString("filename")
			filetype, _ := cmd.Flags().GetString("filetype")
			comment, _ := cmd.Flags().GetString("comment")

			if filename == "" {
				filename = filePath
			}

			respTS, err := apiClient.PostFile(filePath, filename, filetype, comment, username, iconEmoji, thread, currentThreadTS)
			if err != nil {
				return fmt.Errorf("failed to post file: %w", err)
			}
			if thread && currentThreadTS == "" {
				currentThreadTS = respTS
			}
			fmt.Printf("File '%s' posted successfully to profile '%s'.\n", filename, profileName)

		} else {
			// Post from stdin (json)
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				stdinContent, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
				content := string(stdinContent)
				if tee {
					fmt.Print(content)
				}
				respTS, err := apiClient.PostMessage(content, username, iconEmoji, thread, currentThreadTS)
				if err != nil {
					return fmt.Errorf("failed to post message: %w", err)
				}
				if thread && currentThreadTS == "" {
					currentThreadTS = respTS
				}
				fmt.Printf("Message posted successfully to profile '%s'.\n", profileName)
			} else {
				return fmt.Errorf("no file specified and no data from stdin")
			}
		}

		return nil
	},
}

func handleStream(apiClient *client.Client, profileName, overrideUsername, iconEmoji string, thread, tee bool) error {
	fmt.Printf("Starting stream to profile '%s'. Press Ctrl+C to exit.\n", profileName)
	lines := make(chan string)
	scanner := bufio.NewScanner(os.Stdin)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if tee {
				fmt.Println(line)
			}
			lines <- line
		}
		close(lines)
	}()

	var buffer []string
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case line, ok := <-lines:
			if !ok {
				if len(buffer) > 0 {
					fmt.Printf("Flushing %d remaining lines...\n", len(buffer))
					respTS, err := apiClient.PostMessage(strings.Join(buffer, "\n"), overrideUsername, iconEmoji, thread, currentThreadTS)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error flushing remaining lines: %v\n", err)
					}
					if thread && currentThreadTS == "" {
						currentThreadTS = respTS
					}
				}
				fmt.Println("Stream finished.")
				return nil
			}
			buffer = append(buffer, line)
		case <-ticker.C:
			if len(buffer) > 0 {
				respTS, err := apiClient.PostMessage(strings.Join(buffer, "\n"), overrideUsername, iconEmoji, thread, currentThreadTS)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error posting message: %v\n", err)
				}
				if thread && currentThreadTS == "" {
					currentThreadTS = respTS
				}
				fmt.Printf("Posted %d lines to profile '%s'.\n", len(buffer), profileName)
				buffer = nil
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(postCmd)

	postCmd.Flags().StringP("channel", "c", "", "Profile name to use for this post (overrides the default)")
	postCmd.Flags().BoolP("stream", "s", false, "Stream messages from stdin continuously")
	postCmd.Flags().StringP("comment", "m", "", "A comment to post with the file")
	postCmd.Flags().StringP("filename", "n", "", "Filename for the upload")
	postCmd.Flags().String("filetype", "", "Filetype for syntax highlighting")
	postCmd.Flags().StringP("username", "u", "", "Username to post as (overrides the profile default)")
	postCmd.Flags().BoolP("tee", "t", false, "Print stdin to screen before posting")
	postCmd.Flags().Bool("noop", false, "Skip posting to endpoint, for testing purposes")
	postCmd.Flags().StringP("iconemoji", "i", "", "Icon emoji to use for the post (slack provider only)")
	postCmd.Flags().Bool("thread", false, "Post as a reply in a thread (slack provider only)")
}

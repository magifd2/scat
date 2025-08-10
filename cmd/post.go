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

		// Create client
		 apiClient := client.NewClient(profile)

		// Handle stream
		stream, _ := cmd.Flags().GetBool("stream")
		if stream {
			return handleStream(apiClient, profileName)
		}

		// Handle file/stdin post
		var content string
		if len(args) > 0 {
			// Read from file
			filePath := args[0]
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, err)
			}
			content = string(fileContent)
		} else {
			// Read from stdin
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				stdinContent, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
				content = string(stdinContent)
			} else {
				return fmt.Errorf("no file specified and no data from stdin")
			}
		}

		// Post message
		if err := apiClient.PostMessage(content); err != nil {
			return fmt.Errorf("failed to post message: %w", err)
		}

		fmt.Printf("Message posted successfully to profile '%s'.\n", profileName)
		return nil
	},
}

func handleStream(apiClient *client.Client, profileName string) error {
	fmt.Printf("Starting stream to profile '%s'. Press Ctrl+C to exit.\n", profileName)
	lines := make(chan string)
	scanner := bufio.NewScanner(os.Stdin)

	// Read lines from stdin in a separate goroutine
	go func() {
		for scanner.Scan() {
			lines <- scanner.Text()
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
				// Post any remaining lines in the buffer before exiting
				if len(buffer) > 0 {
					fmt.Printf("Flushing %d remaining lines...\n", len(buffer))
					if err := apiClient.PostMessage(strings.Join(buffer, "\n")); err != nil {
						fmt.Fprintf(os.Stderr, "Error flushing remaining lines: %v\n", err)
					}
				}
				fmt.Println("Stream finished.")
				return nil
			}
			buffer = append(buffer, line)
		case <-ticker.C:
			if len(buffer) > 0 {
				if err := apiClient.PostMessage(strings.Join(buffer, "\n")); err != nil {
					fmt.Fprintf(os.Stderr, "Error posting message: %v\n", err)
				}
				fmt.Printf("Posted %d lines to profile '%s'.\n", len(buffer), profileName)
				buffer = nil // Clear buffer after posting
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(postCmd)

	postCmd.Flags().StringP("channel", "c", "", "Profile name to use for this post (overrides the default)")
	postCmd.Flags().BoolP("stream", "s", false, "Stream messages from stdin continuously")
	postCmd.Flags().StringP("comment", "m", "", "A comment to post with the content (not yet implemented)")
	postCmd.Flags().StringP("filename", "n", "", "Filename for the upload (not yet implemented)")
}
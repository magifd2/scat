package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/provider"
	"github.com/spf13/cobra"
)

var postCmd = &cobra.Command{
	Use:   "post [message text]",
	Short: "Post a text message from an argument, file, or stdin",
	Long:  `Posts a text message. The message content is sourced in the following order of precedence: 1. Command-line argument. 2. --from-file flag. 3. Standard input.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Determine profile
		profileName, _ := cmd.Flags().GetString("profile")
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
		channel, _ := cmd.Flags().GetString("channel")
		tee, _ := cmd.Flags().GetBool("tee")
		noop, _ := cmd.Flags().GetBool("noop")
		fromFile, _ := cmd.Flags().GetString("from-file")

		// Get debug flag from persistent flags
		debug, _ := cmd.PersistentFlags().GetBool("debug")

		// Create app context
		appCtx := appcontext.Context{
			Debug: debug,
			NoOp:  noop,
		}

		// Override channel from profile if flag is set
		if channel != "" {
			profile.Channel = channel
		}

		// Get provider instance
		prov, err := GetProvider(appCtx, profile)
		if err != nil {
			return err
		}

		// Handle stream
		stream, _ := cmd.Flags().GetBool("stream")
		if stream {
			// Stream only works with stdin
			if len(args) > 0 || fromFile != "" {
				return fmt.Errorf("cannot use arguments or --from-file with --stream flag")
			}
			return handleStream(prov, profileName, username, iconEmoji, tee)
		}

		// Determine message content from args, file, or stdin
		var content string
		if len(args) > 0 {
			content = strings.Join(args, " ")
		} else if fromFile != "" {
			fileContent, err := os.ReadFile(fromFile)
			if err != nil {
				return fmt.Errorf("failed to read from file %s: %w", fromFile, err)
			}
			content = string(fileContent)
		} else {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				limit := profile.Limits.MaxStdinSizeBytes
				var limitedReader io.Reader = os.Stdin
				if limit > 0 {
					limitedReader = io.LimitReader(os.Stdin, limit+1)
				}
				stdinContent, err := io.ReadAll(limitedReader)
				if err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
				if limit > 0 && int64(len(stdinContent)) > limit {
					return fmt.Errorf("stdin size exceeds the configured limit (%d bytes)", limit)
				}
				content = string(stdinContent)
			} else {
				return fmt.Errorf("no message content provided via argument, --from-file, or stdin")
			}
		}

		if tee && fromFile == "" && len(args) == 0 { // only tee stdin
			fmt.Print(content)
		}

		// Post the message
		if err := prov.PostMessage(content, username, iconEmoji); err != nil {
			return fmt.Errorf("failed to post message: %w", err)
		}
		fmt.Printf("Message posted successfully to profile '%s'.\n", profileName)

		return nil
	},
}

func handleStream(prov provider.Interface, profileName, overrideUsername, iconEmoji string, tee bool) error {
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
					if err := prov.PostMessage(strings.Join(buffer, "\n"), overrideUsername, iconEmoji); err != nil {
						fmt.Fprintf(os.Stderr, "Error flushing remaining lines: %v\n", err)
					}
				}
				fmt.Println("Stream finished.")
				return nil
			}
			buffer = append(buffer, line)
		case <-ticker.C:
			if len(buffer) > 0 {
				if err := prov.PostMessage(strings.Join(buffer, "\n"), overrideUsername, iconEmoji); err != nil {
					fmt.Fprintf(os.Stderr, "Error posting message: %v\n", err)
				}
				buffer = nil
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(postCmd)

	postCmd.Flags().StringP("profile", "p", "", "Profile to use for this post")
	postCmd.Flags().StringP("channel", "c", "", "Override the destination channel for this post")
	postCmd.Flags().String("from-file", "", "Read message body from a file")
	postCmd.Flags().BoolP("stream", "s", false, "Stream messages from stdin continuously")
	postCmd.Flags().BoolP("tee", "t", false, "Print stdin to screen before posting")
	postCmd.Flags().StringP("username", "u", "", "Override the username for this post")
	postCmd.Flags().Bool("noop", false, "Dry run, do not actually post")
	postCmd.Flags().StringP("iconemoji", "i", "", "Icon emoji to use for the post (slack provider only)")
}

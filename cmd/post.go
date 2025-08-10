package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var postCmd = &cobra.Command{
	Use:   "post [file]",
	Short: "Post content from a file or stdin to a configured endpoint",
	Long:  `Posts content to the destination specified in the active profile. Content can be read from a file or from standard input if no file is specified.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Core logic will be implemented here in a later step.
		fmt.Println("Post command is not yet implemented.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(postCmd)

	postCmd.Flags().StringP("channel", "c", "", "Profile name to use for this post (overrides the default)")
	postCmd.Flags().BoolP("stream", "s", false, "Stream messages from stdin continuously")
	postCmd.Flags().StringP("comment", "m", "", "A comment to post with the content")
	postCmd.Flags().StringP("filename", "n", "", "Filename for the upload")
}

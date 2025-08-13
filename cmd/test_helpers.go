
package cmd

import (
	"bytes"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// testExecuteCommandAndCapture is a helper function for testing cobra commands.
// It executes a cobra command and captures its stdout and stderr.
func testExecuteCommandAndCapture(root *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	// Redirect stdout and stderr to buffers
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Execute the command
	root.SetArgs(args)
	err = root.Execute()

	// Restore stdout/stderr and read the captured output
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	stdout = bufOut.String()
	stderr = bufErr.String()
	return
}


package cmd

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

// readPasswordFunc is a variable that holds the function to read a password.
// It can be replaced in tests for mocking purposes.
var readPasswordFunc = term.ReadPassword

// GetPasswordFromPrompt securely reads a password from the terminal.
func GetPasswordFromPrompt(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	tokenBytes, err := readPasswordFunc(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Fprintln(os.Stderr) // Print a newline after password input
	return string(tokenBytes), nil
}

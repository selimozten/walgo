package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// resetCommandFlags resets all flag state in the command tree to prevent
// test pollution from prior executeCommand calls.
func resetCommandFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	for _, child := range cmd.Commands() {
		resetCommandFlags(child)
	}
}

// executeCommand is a helper function for testing cobra commands.
// It resets viper and cobra flag state before each call to prevent test pollution.
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	viper.Reset()
	resetCommandFlags(root)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err = root.Execute()
	return buf.String(), err
}

// captureOutput captures stdout and stderr during function execution
func captureOutput(f func()) (string, string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	outC := make(chan string)
	errC := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outC <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	f()

	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdout := <-outC
	stderr := <-errC

	return stdout, stderr
}

// TestCase represents a test case for command testing
type TestCase struct {
	Name        string
	Args        []string
	ExpectError bool
	Contains    []string
	NotContains []string
	Setup       func()
	Cleanup     func()
}

// runTestCases runs a set of test cases for a command
func runTestCases(t *testing.T, cmd *cobra.Command, cases []TestCase) {
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			// Setup
			if tc.Setup != nil {
				tc.Setup()
			}

			// Execute command
			output, err := executeCommand(cmd, tc.Args...)

			// Check error expectation
			if tc.ExpectError && err == nil {
				t.Errorf("Expected error but got none. Output: %s", output)
			}
			if !tc.ExpectError && err != nil {
				t.Errorf("Unexpected error: %v. Output: %s", err, output)
			}

			// Check output contains expected strings
			for _, expected := range tc.Contains {
				if !bytes.Contains([]byte(output), []byte(expected)) {
					t.Errorf("Output should contain %q, got: %s", expected, output)
				}
			}

			// Check output doesn't contain unexpected strings
			for _, unexpected := range tc.NotContains {
				if bytes.Contains([]byte(output), []byte(unexpected)) {
					t.Errorf("Output should not contain %q, got: %s", unexpected, output)
				}
			}

			// Cleanup
			if tc.Cleanup != nil {
				tc.Cleanup()
			}
		})
	}
}

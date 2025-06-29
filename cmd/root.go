/*
Copyright © 2025 Martin Emde me@martinemde.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"studio-mcp/internal/studio"

	"github.com/spf13/cobra"
)

var debugFlag bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "studio-mcp [--debug] <command> --example \"{{req # required arg}}\" \"[args... # array of args]\"",
	Short: "A tool for running a single command MCP server",
	Long: `studio-mcp is a tool for running a single command MCP server.

  -h, --help - Show this help message and exit.
  --debug - Print debug logs to stderr to diagnose MCP server issues.

the command starts at the first non-flag argument:

  <command> - the shell command to run.
  --example - in the example, this includes the literal argument '--example' is part of the command.
              literal args any shell word.

arguments can be templated as their own shellword or as part of a shellword:

  "{{req # required arg}}" - tell the LLM about a required arg named 'req'.
  "[args... # array of args]" - tell the LLM about an optional array of args named 'args'.
  "[opt # optional string]" - a optional string arg named 'opt' (not in example).
  "https://en.wikipedia.org/wiki/{{wiki_page_name}}" - an example partially templated words.

Example:
  studio-mcp say -v siri "{{speech # a concise phrase to say outloud to the user}}"`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("usage: studio-mcp <command> --example \"{{req # required arg}}\" \"[args... # array of args]\"")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a new Studio instance
		s, err := studio.New(args, debugFlag)
		if err != nil {
			return err
		}

		// Start the MCP server
		return s.Serve()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Define the debug flag
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Print debug logs to stderr to diagnose MCP server issues")
}

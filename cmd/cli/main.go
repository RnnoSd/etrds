package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/RnnoSd/etrds/internal/collect"
	"github.com/RnnoSd/etrds/internal/show"
	"github.com/spf13/cobra"
)

var plain bool

var rootCmd = &cobra.Command{
	Use:   "etrds cmd RDS-uri [query|-F query-file] [-o output-options]",
	Short: "etrds tool querying RDS tool",
	Long:  "etrds still not done",
	Args:  cobra.ArbitraryArgs,
	Run:   run,
}

//var statusCmd = &cobra.Command{
//	Use:   "status [file]",
//	Short: "status",
//	Long:  "status still not done",
//	Args:  cobra.ArbitraryArgs,
//	Run:   status,
//}

var showCmd = &cobra.Command{
	Use:   "show FILE",
	Short: "show",
	Long:  "show still not done",
	Args:  cobra.ArbitraryArgs,
	Run:   show.Run,
}

var collectCmd = &cobra.Command{
	Use:   "collect QUERYNAME",
	Short: "collect",
	Long:  "collect still not done",
	Args:  cobra.ArbitraryArgs,
	Run:   collect.Run,
}

func init() {
	// show subcommand Implementation
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().BoolVarP(&plain, "plain", "p", false, "Display in plain style")

	// collect subcommand Implementation
	rootCmd.AddCommand(collectCmd)
}

func run(cmd *cobra.Command, args []string) {
	fmt.Println("Implment a start-guide and greetings")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}

		os.Exit(1)
	}
}

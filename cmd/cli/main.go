package main

import (
	"fmt"
	"os"
	"bufio"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "etrds cmd RDS-uri [query|-F query-file] [-o output-options]",
	Short: "etrds tool querying RDS tool",
	Long:  "etrds still not done",
	Args:  cobra.ArbitraryArgs,
	Run:   runETRDS,
}

var statusCmd = &cobra.Command{
	Use:   "status [file]",
	Short: "status",
	Long:  "status still not done",
	Args:  cobra.ArbitraryArgs,
	Run:   status,
}

var showCmd = &cobra.Command{
	Use: "show [file]"
	Short: "show",
	Long: "show still not done",
	Args: cobraArbitraryArgs,
	Run: show,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runETRDS(cmd *cobra.Command, args []string) {
	fmt.Println("Command Received")
}

func status(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		entries, err := os.ReadDir("./")
		if err != nil {
			fmt.Printf("config fetch error: %v\n", err)
		}
		for _, file := range entries {
			info, err := file.Info()
			if err != nil {
				continue
			}

			fmt.Printf("%10d bytes %s\n", info.Size(), file.Name())
		}
	}
}

func show(cmd *cobra.Command, args []string) {
	// This function displays queries using batcat
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	hasStdin := false

	for scanner.Scan() {
		hasStdin = true
		lines = append(lines, scanner.Text())
	}
	text = strings.Join(lines, "\n")

	var runCmd *exec.Cmd

	if !hasStdin {
		runCmd = exec.Command("batcat", args...)
	} else {
		execArgs := append(args, text, "-l", "sql")
		runCmd = exec.Command("batcat", execArgs...)
		stdinPipe, err := runCmd.StdinPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creando pipe: %v\n", err)
		}

		go func() {
			defer stdinPipe.Close()
			io.WriteString(stdinPipe, text)
		}()
	}

	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	err := runCmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ejecutando batcat: %v\n", err)
	}

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

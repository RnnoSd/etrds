/*
Package show [LLM-manifest]
*/
package show

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var stdinArgs = []string{
	"-l",
	"sql",
	"--file-name",
	"STDIN sql",
	"--paging=never",
	"--unbuffered",
}

func newBatcatCmd(args ...string) *exec.Cmd {
	cmd := exec.Command("batcat", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

func Run(cmd *cobra.Command, args []string) {
	// [LLM-manifest] implementation summary
	plain, _ := cmd.Flags().GetBool("plain")
	if plain {
		args = append(args, "-p")
		stdinArgs = append(stdinArgs, "-p")
	}
	var err error
	var runBatcat *exec.Cmd

	if len(args) > 0 {
		var argsWOHyphen []string
		hasHyphen := false

		for _, arg := range args {
			if arg != "-" {
				argsWOHyphen = append(argsWOHyphen, arg)
			} else {
				hasHyphen = true
			}
		}

		if len(argsWOHyphen) > 0 {

			runBatcat = newBatcatCmd(argsWOHyphen...)

			err = runBatcat.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error ejecutando batcat: %v\n", err)
			}
		}

		if hasHyphen {
			runBatcat = newBatcatCmd(append([]string{"-"}, stdinArgs...)...)
			runBatcat.Stdin = os.Stdin

			if err := runBatcat.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Wait error: %v\n", err)
			}
		}

	} else {
		runBatcat = newBatcatCmd(stdinArgs...)
		streamFromScanner(runBatcat)
	}
}

func streamFromScanner(run *exec.Cmd) {
	scanner := bufio.NewScanner(os.Stdin)
	stdinPipe, err := run.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, `%v`, err)
	}

	err = run.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ejecutando batcat: %v\n", err)
	}

	go func() {
		defer stdinPipe.Close()
		for scanner.Scan() {
			io.WriteString(stdinPipe, scanner.Text()+"\n")
		}
	}()

	if err := run.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Wait error: %v\n", err)
	}
}

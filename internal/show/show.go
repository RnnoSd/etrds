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

func newBatcatCmd(stdinPipe bool, args ...string) *exec.Cmd {
	cmd := exec.Command("batcat", args...)
	if !stdinPipe {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr

	return cmd
}

func Run(cmd *cobra.Command, args []string) {
	// [LLM-manifest] implementation summary
	plain, _ := cmd.Flags().GetBool("plain")
	if plain {
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
			runBatcat = newBatcatCmd(false, argsWOHyphen...)

			err = runBatcat.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error ejecutando batcat: %v\n", err)
			}
		}

		if hasHyphen {
			runBatcat = newBatcatCmd(false, append([]string{"-"}, stdinArgs...)...)
			runBatcat.Stdin = os.Stdin

			if err := runBatcat.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "error de espera: %v\n", err)
			}
		}

	} else {
		runBatcat = newBatcatCmd(true, stdinArgs...)
		err := streamFromScanner(cmd, runBatcat)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error de espera: %v\n", err)
		}
	}
}

func streamFromScanner(cmd *cobra.Command, run *exec.Cmd) error {
	scanner := bufio.NewScanner(os.Stdin)
	stdinPipe, err := run.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, `%v`, err)
		return err
	}
	stdoutPipe, err := run.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, `%v`, err)
		return err
	}

	err = run.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error ejecutando batcat: %v\n", err)
		return err
	}

	go func() {
		defer stdinPipe.Close()
		for scanner.Scan() {
			io.WriteString(stdinPipe, scanner.Text()+"\n")
		}
	}()

	cobraStdout := cmd.OutOrStdout()
	io.Copy(cobraStdout, stdoutPipe)

	if f, ok := cobraStdout.(*os.File); ok {
		_ = f.Sync()
	}
	if err := run.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "error de espera: %v\n", err)
		return err
	}

	return nil
}

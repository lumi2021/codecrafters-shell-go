package main

import (
	"bufio"
	"fmt"
	"os"
	"shell/app/execution"
	"shell/app/parser"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("$ ")

		commandStr, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println()
				break
			}

			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		command, err := parser.ParseCommand(commandStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		execution.Execute(command)
	}
}

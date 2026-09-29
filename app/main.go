package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var _ = fmt.Print

func main() {
	for true {
		fmt.Print("$ ")

		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		command = strings.TrimSpace(command)

		if command == "exit" {
			break
		}

		if strings.HasPrefix(command, "echo ") {
			fmt.Println(command[:5]);
		}

		fmt.Println(command+ ": command not found")
	}
}

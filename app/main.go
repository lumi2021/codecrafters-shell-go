package main

import (
	"strings"
	"bufio"
	"fmt"
	"os"
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
		command = strings.Trim(command)

		if (command == "exit") os.Exit(1)

		fmt.Println(command[:len(command)-1] + ": command not found")
	}
}

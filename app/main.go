package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var _ = fmt.Print

func main() {
execloop:
	for true {
		fmt.Print("$ ")

		commandStr, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		command := strings.Fields(commandStr)

		switch command[0] {
		case "echo":
			fmt.Println(strings.TrimSpace(commandStr)[5:])

		case "exit":
			break execloop

		case "type":
			switch command[1] {
			case "echo":
			case "exit":
			case "type":
				fmt.Println(command[1] + " is a shell builtin")
			}

		default:
			fmt.Println(command[1] + ": not found")
		}

		fmt.Println(command[0] + ": command not found")
	}
}

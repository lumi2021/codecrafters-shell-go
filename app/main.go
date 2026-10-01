package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
			case "echo",
				"exit",
				"type":
				fmt.Println(command[1] + " is a shell builtin")

			default:
				path, success := searchForExec(command[1], getEnvPath())
				if !success {
					fmt.Println(command[1] + ": not found")
					continue
				}
				fmt.Println(command[1] + " is " + path)
			}

		default:
			path, success := searchForExec(command[0], getEnvPath())
			if !success {
				fmt.Println(command[0] + ": command not found")
				continue
			}

			execute(path, command[1:])
		}
	}
}

func getEnvPath() []string {
	path := os.Getenv("PATH")
	paths := strings.Split(path, string(os.PathListSeparator))
	return paths
}

func searchForExec(exec string, paths []string) (string, bool) {
	for _, dir := range paths {
		path := filepath.Join(dir, exec)

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			continue
		}

		if info.IsDir() {
			continue
		}

		if info.Mode().Perm()&0111 == 0 {
			continue
		}

		return path, true
	}

	return "", false
}

func execute(execPath string, args []string) {
	cmd := exec.Command(execPath, "-la")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
}

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
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
		command := splitArgs(commandStr)

		switch command[0] {
		case "echo":
			fmt.Println(strings.Join(command[1:], " "))

		case "exit":
			break execloop

		case "type":
			switch command[1] {
			case "echo",
				"exit",
				"type",
				"pwd",
				"cd":
				fmt.Println(command[1] + " is a shell builtin")

			default:
				path, success := searchForExec(command[1], getEnvPath())
				if !success {
					fmt.Println(command[1] + ": not found")
					continue
				}
				fmt.Println(command[1] + " is " + path)
			}

		case "pwd":
			fmt.Println(getCwd())

		case "cd":
			if len(command) < 2 {
				panic("cd: missing argument")
			}
			err := cd(command[1])
			if err != nil {
				fmt.Println("cd: " + command[1] + ": No such file or directory")
			}

		default:
			path, success := searchForExec(command[0], getEnvPath())
			if !success {
				fmt.Println(command[0] + ": command not found")
				continue
			}

			execute(path, command)
		}
	}
}

func splitArgs(cmdline string) []string {
	var args []string

	var arg strings.Builder
	var parenthesis = '\x00'

	runes := []rune(cmdline)
	for i := 0; i < len(runes); {

		r, size := utf8.DecodeRuneInString(cmdline[i:])
		i += size

		if r == '\\' {
			if i >= len(cmdline) {
				arg.WriteRune('\\')
				break
			}

			escaped, escapedSize := utf8.DecodeRuneInString(cmdline[i:])
			i += escapedSize

			switch escaped {
			case 'n':
				arg.WriteRune('\n')
			case 'r':
				arg.WriteRune('\r')
			case 't':
				arg.WriteRune('\t')
			case '\\':
				arg.WriteRune('\\')
			case '\'':
				arg.WriteRune('\'')
			case '"':
				arg.WriteRune('"')
			default:
				arg.WriteRune(escaped)
			}

			continue
		}

		if (parenthesis == '\x00' || parenthesis == r) && (r == '\'' || r == '"') {
			if parenthesis == '\x00' {
				parenthesis = r
			} else {
				parenthesis = '\x00'
			}
			continue
		}

		if parenthesis == '\x00' && unicode.IsSpace(r) {
			if arg.Len() != 0 {
				args = append(args, arg.String())
				arg.Reset()
			}
			continue
		}

		arg.WriteRune(r)
	}

	if arg.Len() > 0 {
		args = append(args, arg.String())
	}

	return args
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
	cmd := exec.Command(execPath, args[1:]...)
	cmd.Args[0] = args[0]

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
}

func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd
}
func cd(to string) error {
	if to == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		to = home
	}

	return os.Chdir(to)
}

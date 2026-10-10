package execution

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"shell/app/parser"
	"strings"
)

type ShellIO struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type BuiltinFunc func(args []string, io ShellIO) error

var builtins map[string]BuiltinFunc

func init() {
	builtins = map[string]BuiltinFunc{
		"echo": builtinEcho,
		"pwd":  builtinPwd,
		"cd":   builtinCd,
		"type": builtinType,
		"exit": builtinExit,
	}
}

func Execute(command *parser.Command) error {
	shellIO := ShellIO{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	var files []*os.File

	for _, redirect := range command.Redirects {
		var (
			file *os.File
			err  error
		)

		switch redirect.FD {
		case 0:
			file, err = os.Open(redirect.Path)

		case 1, 2:
			file, err = os.OpenFile(
				redirect.Path,
				redirect.Mode,
				0666,
			)

		default:
			return fmt.Errorf(
				"unsupported file descriptor: %d",
				redirect.FD,
			)
		}

		if err != nil {
			for _, f := range files {
				_ = f.Close()
			}
			return err
		}

		files = append(files, file)

		switch redirect.FD {
		case 0:
			shellIO.Stdin = file
		case 1:
			shellIO.Stdout = file
		case 2:
			shellIO.Stderr = file
		}
	}

	defer func() {
		for _, file := range files {
			_ = file.Close()
		}
	}()

	args := command.Args

	if len(args) == 0 {
		return nil
	}

	if builtin, exists := builtins[args[0]]; exists {
		return builtin(args, shellIO)
	}

	path, exists := searchForExec(args[0], getEnvPath())

	if !exists {
		fmt.Fprintln(
			shellIO.Stderr,
			args[0]+": command not found",
		)

		return nil
	}

	return callExecutable(path, args, shellIO)
}

func builtinExit(args []string, io ShellIO) error {
	os.Exit(0)
	return nil
}

func builtinEcho(args []string, io ShellIO) error {
	fmt.Fprintln(io.Stdout, strings.Join(args[1:], " "))
	return nil
}

func builtinType(args []string, io ShellIO) error {
	if len(args) < 2 {
		return nil
	}

	name := args[1]

	if _, exists := builtins[name]; exists {
		fmt.Fprintln(io.Stdout, name+" is a shell builtin")
		return nil
	}

	path, exists := searchForExec(name, getEnvPath())

	if !exists {
		fmt.Fprintln(io.Stdout, name+": not found")
		return nil
	}

	fmt.Fprintln(io.Stdout, name+" is "+path)
	return nil
}

func builtinPwd(args []string, io ShellIO) error {
	fmt.Fprintln(io.Stdout, getCwd())
	return nil
}

func builtinCd(args []string, io ShellIO) error {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "cd: missing argument")
		return nil
	}

	if err := cd(args[1]); err != nil {
		fmt.Println("cd: " + args[1] + ": No such file or directory")
	}

	return nil
}

func callExecutable(path string, args []string, io ShellIO) error {
	cmd := exec.Command(path, args[1:]...)
	cmd.Args[0] = args[0]

	cmd.Stdin = io.Stdin
	cmd.Stdout = io.Stdout
	cmd.Stderr = io.Stderr

	return cmd.Run()
}

func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return cwd
}

func getEnvPath() []string {
	path := os.Getenv("PATH")

	return strings.Split(
		path,
		string(os.PathListSeparator),
	)
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

func searchForExec(executable string, paths []string) (string, bool) {
	for _, dir := range paths {
		path := filepath.Join(dir, executable)

		info, err := os.Stat(path)
		if err != nil {
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

package main

import (
	"bufio"
	"fmt"
	"os"
)

var _ = fmt.Print

func main() {
	fmt.Print("$ ")

	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
	fmt.Println(command[:len(command)-1] + ": command not found")
}

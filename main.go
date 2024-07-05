package main

import (
	"bufio"
	"fmt"
	"os"
)

type Command string

const (
	cmdConfig   Command = "config"
	cmdAdd      Command = "add"
	cmdLog      Command = "log"
	cmdCommit   Command = "commit"
	cmdCheckout Command = "checkout"
)

const (
	vcsDir     = "./vcs/"
	configFile = vcsDir + "config.txt"
	indexFile  = vcsDir + "index.txt"
)

var allowedCommands = []Command{
	cmdConfig,
	cmdAdd,
	cmdLog,
	cmdCommit,
	cmdCheckout,
}

var helpMsg = map[Command]string{
	cmdConfig:   "Get and set a username.",
	cmdAdd:      "Add a file to the index.",
	cmdLog:      "Show commit logs.",
	cmdCommit:   "Save changes.",
	cmdCheckout: "Restore a file.",
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" {
		displayHelp()
		return
	}

	var cmd = Command(os.Args[1])
	var args = os.Args[2:]

	switch cmd {
	case cmdConfig:
		config(args)
	case cmdAdd:
		add(args)
	case cmdLog:
		log(args)
	case cmdCommit:
		commit(args)
	case cmdCheckout:
		checkout(args)
	default:
		fmt.Printf("'%s' is not a SVCS command.\n", cmd)
	}
}

func displayHelp() {
	fmt.Println("These are SVCS commands:")
	for _, command := range allowedCommands {
		fmt.Printf("%-8s  %s\n", command, helpMsg[command])
	}
}

func config(args []string) {
	if len(args) == 1 {
		setUsername(args[0])
	}

	if username, err := getUsername(); err == nil {
		fmt.Printf("The username is %s.\n", username)
	} else {
		fmt.Println("Please, tell me who you are.")
	}
}

func setUsername(username string) {
	os.MkdirAll(vcsDir, os.ModePerm)

	file, err := os.Create(configFile)
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(username)
}

func getUsername() (string, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	return scanner.Text(), scanner.Err()
}

func add(args []string) {
	if len(args) != 1 {
		displayTracking()
		return
	}

	os.MkdirAll(vcsDir, os.ModePerm)

	filename := args[0]

	if _, err := os.Stat(filename); err != nil {
		fmt.Printf("Can't find '%s'.\n", filename)
		return
	}

	file, err := os.OpenFile(indexFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString(filename + "\n")

	fmt.Printf("The file '%s' is tracked.", filename)
}

func displayTracking() {
	file, err := os.Open(indexFile)
	if err != nil {
		fmt.Println("Add a file to the index.")
		return
	}
	defer file.Close()

	fmt.Println("Tracked files:")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func log(args []string) {
	// todo: stub
}

func commit(args []string) {
	// todo: stub
}

func checkout(args []string) {
	// todo: stub
}

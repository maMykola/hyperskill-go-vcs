package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Command string

type Commit struct {
	Hash     string
	Username string
	Message  string
}

const (
	cmdConfig   Command = "config"
	cmdAdd      Command = "add"
	cmdLog      Command = "log"
	cmdCommit   Command = "commit"
	cmdCheckout Command = "checkout"
)

const (
	vcsDir     = "./vcs/"
	commitsDir = vcsDir + "commits"
	configFile = vcsDir + "config.txt"
	indexFile  = vcsDir + "index.txt"
	logFile    = vcsDir + "log.txt"
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
		doConfig(args)
	case cmdAdd:
		doAdd(args)
	case cmdLog:
		doLog()
	case cmdCommit:
		doCommit(args)
	case cmdCheckout:
		doCheckout(args)
	default:
		fmt.Printf("'%s' is not a SVCS command.\n", cmd)
	}
}

func (c *Commit) show() {
	fmt.Println("commit", c.Hash)
	fmt.Println("Author:", c.Username)
	fmt.Println(c.Message)
}

func displayHelp() {
	fmt.Println("These are SVCS commands:")
	for _, command := range allowedCommands {
		fmt.Printf("%-8s  %s\n", command, helpMsg[command])
	}
}

func doConfig(args []string) {
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

func doAdd(args []string) {
	if len(args) != 1 {
		showTrackedFiles()
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
		log.Fatal(err)
	}
	defer file.Close()

	file.WriteString(filename + "\n")

	fmt.Printf("The file '%s' is tracked.\n", filename)
}

func showTrackedFiles() {
	files := getTrackedFiles()
	if len(files) == 0 {
		fmt.Println("Add a file to the index.")
		return
	}

	fmt.Println("Tracked files:")
	for _, filename := range files {
		fmt.Println(filename)
	}
}

func getTrackedFiles() []string {
	file, err := os.Open(indexFile)
	if err != nil {
		return nil
	}
	defer file.Close()

	files := make([]string, 0, 5)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		files = append(files, scanner.Text())
	}

	return files
}

func doLog() {
	commits := getCommits()
	if len(commits) == 0 {
		fmt.Println("No commits yet.")
		return
	}

	i := len(commits) - 1
	commits[i].show()

	for i--; i >= 0; i-- {
		fmt.Println()
		commits[i].show()
	}
}

func getCommits() []Commit {
	file, err := os.Open(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		log.Fatal(err)
	}
	defer file.Close()

	var commits []Commit

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		data := strings.SplitN(line, " ", 3)
		commits = append(commits, Commit{
			Hash:     data[0],
			Username: data[1],
			Message:  data[2],
		})
	}

	return commits
}

func findCommit(hash string) (Commit, bool) {
	for _, commit := range getCommits() {
		if commit.Hash == hash {
			return commit, true
		}
	}

	return Commit{}, false
}

func restoreCommit(commit Commit) {
	commit.show()
}

func doCommit(args []string) {
	if len(args) != 1 {
		fmt.Println("Message was not passed.")
		return
	}

	username, err := getUsername()
	if err != nil {
		fmt.Println("Please, tell me who are you.")
		return
	}

	if saveCommit(username, args[0]) {
		fmt.Println("Changes are committed.")
	} else {
		fmt.Println("Nothing to commit.")
	}
}

func doCheckout(args []string) {
	if len(args) != 1 {
		fmt.Println("Commit id was not passed.")
		return
	}

	commit, ok := findCommit(args[0])
	if !ok {
		fmt.Println("Commit does not exist.")
		return
	}

	restoreCommit(commit)
}

func saveCommit(username, message string) (saved bool) {
	files := getTrackedFiles()
	hash := computeHash(files)
	if hasChanges(hash) {
		commit := Commit{
			Hash:     hash,
			Username: username,
			Message:  message,
		}

		commitFiles(hash, files)
		addLog(commit)

		saved = true
	}
	return
}

func computeHash(files []string) string {
	sha256Hash := sha256.New()
	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := io.Copy(sha256Hash, file); err != nil {
			log.Fatal(err)
		}

		file.Close()
	}

	return hex.EncodeToString(sha256Hash.Sum(nil))
}

func hasChanges(hash string) bool {
	path := filepath.Join(commitsDir, hash)

	if _, err := os.Stat(path); err == nil {
		return false
	}

	return true
}

func commitFiles(hash string, files []string) {
	commitPath := filepath.Join(commitsDir, hash)
	for _, filename := range files {
		copyFile(filepath.Join(commitPath, filename), filename)
	}
}

func copyFile(dest, src string) {
	os.MkdirAll(filepath.Dir(dest), os.ModePerm)

	source, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()

	file, err := os.Create(dest)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if _, err := io.Copy(file, source); err != nil {
		log.Fatal(err)
	}
}

func addLog(c Commit) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("%s %s %s\n", c.Hash, c.Username, c.Message))
}

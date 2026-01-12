package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func IsPresent() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func Init(dir string) error {
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}
	return ensureUserConfig(dir)
}

func ensureUserConfig(dir string) error {
	cmd := exec.Command("git", "config", "user.email")
	cmd.Dir = dir
	if err := cmd.Run(); err == nil {
		return nil
	}

	fmt.Println("Git user identity not configured.")
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)

	fmt.Print("Enter your email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	email = strings.TrimSpace(email)

	if err := setConfig(dir, "user.name", name); err != nil {
		return err
	}
	return setConfig(dir, "user.email", email)
}

func setConfig(dir, key, value string) error {
	cmd := exec.Command("git", "config", key, value)
	cmd.Dir = dir
	return cmd.Run()
}

func Commit(dir string, message string) error {
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = dir
	output, err := statusCmd.Output()
	if err != nil {
		return err
	}
	if len(output) == 0 {
		return nil
	}

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = dir
	if err := addCmd.Run(); err != nil {
		return err
	}

	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = dir
	return commitCmd.Run()
}

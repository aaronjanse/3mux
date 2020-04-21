package pane

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

func getShellPath() (string, error) {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell, nil
	}

	username := os.Getenv("USER")

	file, err := os.Open("/etc/passwd")
	if err != nil {
		return "", err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if parts[0] == username {
			return parts[len(parts)-1], nil
		}
	}
	return "", errors.New("Could not find shell to use")
}

package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func printRofiError(message string) {
	command := exec.Command("rofi", "-dmenu", "-p", rofiPrompt, "-sep", "\n")
	command.Stdin = strings.NewReader(fmt.Sprintf("Error:\n %s", message))
	command.Run()
}

func openInObsidian(vaultName, filePath string) {
	var cmd *exec.Cmd

	if filePath == "" {
		cmd = exec.Command(obsidianAppExec, fmt.Sprintf("obsidian://open?vault=%s", vaultName))
	} else {
		cmd = exec.Command(obsidianAppExec, fmt.Sprintf("obsidian://open?vault=%s&file=%s", vaultName, filePath))
	}

	cmd.Start()
}

func extractSelection(input string) (name string, path string) {
	parts := strings.Split(input, "|")
	if len(parts) != 2 {
		return "", ""
	}

	nameRe := regexp.MustCompile(`<b>(.*?)</b>`)
	matchName := nameRe.FindStringSubmatch(parts[0])
	if len(matchName) > 1 {
		name = matchName[1]
	}

	path = strings.TrimSpace(parts[1])
	if idx := strings.Index(path, "</span>"); idx != 1 {
		path = path[:idx]
	}

	return name, path
}

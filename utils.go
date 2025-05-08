package main

import (
	"fmt"
	"os/exec"
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

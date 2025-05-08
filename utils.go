package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
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

func extractVaultSelection(input string) (name, path string) {
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

func extractFileSelection(input string) (name string) {
	nameRe := regexp.MustCompile(`<span>(.*?)</span>`)
	matchName := nameRe.FindStringSubmatch(input)
	if len(matchName) > 1 {
		name = matchName[1]
	}
	return name
}

// check for obsidian supported media files
func getFileIcon(filename string) string {
	audioExtensions := map[string]bool{
		// audio
		".mp3": true, ".wav": true, ".ogg": true,
		".flac": true, ".m4a": true, ".webm": true,
		".3gp": true,
	}
	videoExtensions := map[string]bool{
		// video
		".mp4": true, ".mkv": true, ".mov": true,
		".ogv": true,
	}

	imageExtensions := map[string]bool{
		// images
		".avif": true, ".bmp": true, ".gif": true,
		".jpeg": true, "jpg": true, ".png": true,
		".svg": true, ".webp": true,
	}
	pdfExtensions := map[string]bool{
		// pdf
		".pdf": true,
	}

	markdownExtensions := map[string]bool{
		".md": true,
	}

	fileExtension := filepath.Ext(strings.ToLower(filename))

	// check
	if audioExtensions[fileExtension] {
		return "\uf1c7"
	} else if videoExtensions[fileExtension] {
		return "\uf1c8"
	} else if imageExtensions[fileExtension] {
		return "\uf1c5"
	} else if pdfExtensions[fileExtension] {
		return "\uf1c1"
	} else if markdownExtensions[fileExtension] {
		return "\uebaf"
	} else {
		return ""
	}
}

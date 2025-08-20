package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// print error in rofi
func printRofiError(message string) {
	command := exec.Command("rofi", "-e", message)
	command.Run()
}

// open the selected vault or file in obsidian
func openInObsidian(vaultName, filePath string) {
	var cmd *exec.Cmd

	if filePath == "" {
		cmd = exec.Command(obsidianAppExec, fmt.Sprintf("obsidian://open?vault=%s", vaultName))
	} else {
		cmd = exec.Command(obsidianAppExec, fmt.Sprintf("obsidian://open?vault=%s&file=%s", vaultName, filePath))
	}

	cmd.Start()
}

// extract file name and path from selection
// format: <b>%s</b> | <span>%s</span>
func extractSelection(input string) (name, path string) {
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

// extract file name from the selection
// format: <span>%s</span>
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
		".jpeg": true, ".jpg": true, ".png": true,
		".svg": true, ".webp": true,
	}

	otherExtensions := map[string]string{
		".pdf":    "\uf1c1",
		".canvas": "\\udb80\\ude15",
		".md":     "\uebaf",
	}

	fileExtension := filepath.Ext(strings.ToLower(filename))

	// check
	var fileIcon string

	if fileExtension == ".md" && strings.Contains(strings.ToLower(filename), ".excalidraw") {
		fileIcon = "\uee75"
	} else if audioExtensions[fileExtension] {
		fileIcon = "\uf1c7"
	} else if videoExtensions[fileExtension] {
		fileIcon = "\uf1c8"
	} else if imageExtensions[fileExtension] {
		fileIcon = "\uf1c5"
	} else if icon, ok := otherExtensions[fileExtension]; ok {
		fileIcon = icon
	} else {
		fileIcon = ""
	}

	if len(fileIcon) > 6 {
		icon, err := convertSurrogatePair(fileIcon)
		if err != nil {
			return fileIcon
		}
		return fmt.Sprintf("%c", icon)
	} else {
		return fileIcon
	}
}

// converts a surrogate pair to a rune
// ref: https://datacadamia.com/data/type/text/surrogate
func convertSurrogatePair(s string) (rune, error) {
	// remove \u
	s = strings.ReplaceAll(s, "\\u", "")
	if len(s) != 8 {
		return 0, fmt.Errorf("invalid surrogate pair length")
	}

	// split high and low codepoint
	highHex := s[:4]
	lowHex := s[4:]

	// convert to hex
	lead, err := strconv.ParseUint(highHex, 16, 16)
	if err != nil {
		return 0, err
	}
	trail, err := strconv.ParseUint(lowHex, 16, 16)
	if err != nil {
		return 0, err
	}

	// convert into actual code point
	offset := 0x10000 - (0xD800 << 10) - 0xDC00
	codepoint := (lead << 10) + trail + uint64(offset)
	return rune(codepoint), nil
}

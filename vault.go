package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ConfigData struct {
	Vaults map[string]struct {
		Path string `json:"path"`
	} `json:"vaults"`
}

type Vault struct {
	Path string
	Name string
}

// Get vaults path and name from config file
func getVaultsFromConfig(configFilePath string) ([]Vault, error) {
	file, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var data ConfigData
	if err := json.Unmarshal(file, &data); err != nil {
		return nil, err
	}

	var vaults []Vault

	for _, info := range data.Vaults {
		path := info.Path
		name := filepath.Base(path)

		vaults = append(vaults, Vault{
			Path: path,
			Name: name,
		})
	}

	if len(vaults) <= 0 {
		return nil, fmt.Errorf("No vaults found.")
	}

	return vaults, nil
}

// Select Vaults in Rofi Menu
func selectVault(vaults []Vault) *Vault {
	var options strings.Builder

	for _, vault := range vaults {
		options.WriteString(fmt.Sprintf("\ueb29 <b>%s</b><span foreground='#D3D3D3'> | %s</span>\n", vault.Name, vault.Path))
	}

	command := exec.Command("rofi", "-dmenu", "-markup-rows", "-i", "-p", rofiPrompt, "-sep", "\n")
	command.Stdin = strings.NewReader(options.String())

	out, err := command.Output()
	if err != nil {
		return nil
	}

	userSelection := strings.TrimSpace(string(out))
	if userSelection == "" {
		return nil
	}

	vaultName, vaultPath := extractVaultSelection(userSelection)

	return &Vault{
		Path: strings.TrimSpace(vaultPath),
		Name: strings.TrimSpace(vaultName),
	}
}

// Browse files
func browseVault(vaultPath string) {
	vaultName := filepath.Base(vaultPath)
	currentPath := vaultPath

	for {
		dir, err := os.ReadDir(currentPath)
		if err != nil {
			printRofiError(fmt.Sprintf("Error reading directory: %s", err))
			return
		}

		var options strings.Builder
		options.WriteString("\uedf5 <span>Open</span> vault\n")
		options.WriteString("\uf002 <span>Search</span> files\n")
		if currentPath != vaultPath {
			options.WriteString("<span>../</span>\n")
		}

		for _, entry := range dir {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			filename := entry.Name()
			fileIcon := getFileIcon(filename)

			if entry.IsDir() {
				filename = fmt.Sprintf("\uf07b <span>%s/</span>", filename)
			} else {
				filename = fmt.Sprintf("%s <span>%s</span>", fileIcon, filename)
			}
			options.WriteString(fmt.Sprintf("%s\n", filename))
		}

		command := exec.Command("rofi", "-dmenu", "-markup-rows", "-i", "-p", fmt.Sprintf("Browsing"), "-sep", "\n")
		command.Stdin = strings.NewReader(options.String())
		out, err := command.Output()
		if err != nil {
			return
		}

		userSelection := extractFileSelection(string(out))
		if strings.ToLower(userSelection) == "open" {
			openInObsidian(vaultName, "")
			break
		}
		if strings.ToLower(userSelection) == "search" {
			if results := searchFiles(vaultPath); results != nil {
				selectedFile := displaySearchResults(results)
				if selectedFile != "" {
					relPath, _ := filepath.Rel(vaultPath, selectedFile)
					openInObsidian(vaultName, relPath)
				}
			}
			continue
		}
		currentPath = filepath.Join(currentPath, userSelection)

		if info, err := os.Stat(currentPath); err == nil && !info.IsDir() {
			path := strings.Replace(currentPath, vaultPath, "", 1)
			openInObsidian(vaultName, path)
			break
		}
	}
}

// searchFiles prompts for a search term and returns matching files
func searchFiles(vaultPath string) []string {
	// Prompt for search term
	command := exec.Command("rofi", "-dmenu", "-markup-rows", "-i", "-p", "Search")
	command.Stdin = strings.NewReader("")
	out, err := command.Output()
	if err != nil {
		return nil
	}

	searchTerm := strings.TrimSpace(string(out))
	if searchTerm == "" {
		return nil
	}

	var results []string
	err = filepath.Walk(vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(strings.ToLower(info.Name()), strings.ToLower(searchTerm)) {
			results = append(results, path)
		}
		return nil
	})

	if err != nil {
		printRofiError(fmt.Sprintf("Error searching files: %s", err))
		return nil
	}

	return results
}

// displaySearchResults shows search results in rofi and returns the selected file
func displaySearchResults(results []string) string {
	if len(results) == 0 {
		printRofiError("No files found")
		return ""
	}

	var options strings.Builder
	for _, path := range results {
		filename := filepath.Base(path)
		fileIcon := getFileIcon(filename)
		if fileIcon == "" {
			fileIcon = "\uf15b" // default file icon
		}
		options.WriteString(fmt.Sprintf("%s <b>%s</b><span foreground='#D3D3D3'> | %s</span>\n", 
			fileIcon, filename, path))
	}

	command := exec.Command("rofi", "-dmenu", "-markup-rows", "-i", "-p", "Results")
	command.Stdin = strings.NewReader(options.String())
	out, err := command.Output()
	if err != nil {
		return ""
	}

	selection := strings.TrimSpace(string(out))
	if selection == "" {
		return ""
	}

	_, path := extractVaultSelection(selection)
	return path
}

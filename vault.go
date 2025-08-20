package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

type File Vault

var vaultFiles = make(map[string][]File)

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

	// sort vault alphabetically
	sort.Slice(vaults, func(i, j int) bool {
		return strings.ToLower(vaults[i].Name) < strings.ToLower(vaults[j].Name)
	})

	return vaults, nil
}

// Select vaults in Rofi menu
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

	vaultName, vaultPath := extractSelection(userSelection)

	return &Vault{
		Path: strings.TrimSpace(vaultPath),
		Name: strings.TrimSpace(vaultName),
	}
}

// fetch all the files from the vault
func fetchVaultFiles(vaultPath string) {
	var files []File

	filepath.Walk(vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		fileName := info.Name()

		// skip hidden files and directories
		hidden := strings.HasPrefix(fileName, ".")
		if info.IsDir() && hidden {
			return filepath.SkipDir
		}

		// add files path and name
		if !info.IsDir() && !hidden {
			files = append(files, File{
				Path: path,
				Name: fileName,
			})
		}
		return nil
	})

	vaultFiles[vaultPath] = files
}

// Browse vault
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
		options.WriteString("\uedf5   <span>Open</span> vault\n")
		options.WriteString("\uf002  <span>Search</span> files\n")
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
				filename = fmt.Sprintf("\uf07b   <span>%s/</span>", filename)
			} else {
				filename = fmt.Sprintf("%s   <span>%s</span>", fileIcon, filename)
			}
			options.WriteString(fmt.Sprintf("%s\n", filename))
		}

		command := exec.Command("rofi", "-dmenu", "-markup-rows", "-i", "-p", "Browsing", "-sep", "\n")
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
			selectedFile := searchFiles(vaultPath)
			path := strings.Replace(selectedFile, vaultPath, "", 1)
			if path != "" {
				openInObsidian(vaultName, path)
			}
			break
		}
		currentPath = filepath.Join(currentPath, userSelection)

		if info, err := os.Stat(currentPath); err == nil && !info.IsDir() {
			path := strings.Replace(currentPath, vaultPath, "", 1)
			openInObsidian(vaultName, path)
			break
		}
	}
}

// search files and return path
func searchFiles(vaultPath string) string {
	command := exec.Command("rofi", "-dmenu", "-markup-rows", "-i", "-p", "Search")
	var options strings.Builder

	files := vaultFiles[vaultPath]
	for _, file := range files {
		fileIcon := getFileIcon(file.Name)
		options.WriteString(fmt.Sprintf("%s <b>%s</b><span foreground='#D3D3D3'> | %s</span>\n",
			fileIcon, file.Name, file.Path))
	}

	if options.String() == "" {
		printRofiError("No files found")
		return ""
	}

	command.Stdin = strings.NewReader(options.String())
	out, err := command.Output()
	if err != nil {
		return ""
	}

	selection := strings.TrimSpace(string(out))
	if selection == "" {
		return ""
	}

	_, path := extractSelection(selection)
	return path
}

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
		options.WriteString(fmt.Sprintf("<b>%s</b><span foreground='#D3D3D3'> | %s</span>\n", vault.Name, vault.Path))
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

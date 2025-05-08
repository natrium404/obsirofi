package main

import (
	"os"
	"path/filepath"
)

const (
	obsidianConfigDir = "$HOME/.config/obsidian/"
	obsidianAppExec   = "obsidian"
	rofiPrompt        = "Obsidian Vault"
)

func main() {
	configFilePath := os.ExpandEnv(filepath.Join(obsidianConfigDir, "obsidian.json"))
	vaults, err := getVaultsFromConfig(configFilePath)
	if err != nil {
		printRofiError(err.Error())
		return
	}

	selectedVault := selectVault(vaults)
	if selectedVault == nil {
		return
	}

	openInObsidian(selectedVault.Name, "")
}

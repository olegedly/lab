package main

import (
	"errors"
	"os"
)

func getArgs() []string {
	return os.Args
}

func errUsage() error {
	return errors.New("usage: obsidian-omnisearch <obsidian_vault_path>")
}

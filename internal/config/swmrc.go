package config

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	swmDir = "swm"
	swmrcFile = "swmrc"
)

// FindAndRunSwmrc finds first existing swmrc file and executes it
// Searched locations:
//     - if customPath is provided, only customPath is tried
//     1) {XDG_CONFIG_HOME}/{swmDir}/{swmrcFile}
//     2) {HOME}/.config/{swmDir}/{swmrcFile}
//     3) {HOME}/.{swmDir}/{swmrcFile}
func FindAndRunSwmrc(customPath string) {
	log.Printf("Trying to execute config")

	if customPath != "" {
		path := customPath
		if path[0] != '/' {
			currentDir, err := os.Getwd()
			if err != nil {
				log.Printf("Cannot get current working directory: %s", err)
			}
			path = filepath.Join(currentDir, customPath)
		}
		if _, err := os.Stat(path); err != nil {
			log.Printf("Provided config file does not seem to exist: %s", err)
		} else {
			executeConfig(path)
		}
		return
	}

	configDir := os.Getenv("XDG_CONFIG_HOME")
	homeDir := os.Getenv("HOME")
	var files []string

	if configDir != "" {
		files = append(
			files,
			filepath.Join(configDir, swmDir, swmrcFile),
		)
	}
	if homeDir != "" {
		files = append(
			files,
			filepath.Join(homeDir, ".config", swmDir, swmrcFile),
			filepath.Join(homeDir, fmt.Sprintf(".%s", swmDir), swmrcFile),
		)
	}

	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			executeConfig(file)
			return
		}
	}

	log.Printf("No config file found, searched locations:")
	for _, file := range files {
		log.Printf("\t%s", file)
	}
}

func executeConfig(file string) {
	log.Printf("Executing config file \"%s\"", file)

	err := exec.Command(file).Run()

	if err != nil {
		log.Printf("Error executing config file: %s", err)
	}
}

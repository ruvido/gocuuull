package main

import (
   "fmt"
    "os"
    "os/exec"
    "strings"
    "path/filepath"
    // "errors"
)

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func fileExists(filePath string) bool {
    _, err := os.Stat(filePath)
    if os.IsNotExist(err) {
        return false
    }
    return err == nil
}

func touch(filePath string) error {
    cmd := exec.Command("touch", filePath)
    return cmd.Run()
}

func buildPath(filename, library, newlibrary, suffix string) string {
	// Check if filename starts with library path
	if !strings.HasPrefix(filename, library) {
		fmt.Println("Error: filename does not start with the specified library path")
		return ""
	}

	// Remove the library prefix from the filename
	relativePath := strings.TrimPrefix(filename, library)
	
	// Construct the new path with the new library
	newPath := filepath.Join(newlibrary, relativePath)

	// If a suffix is provided, append it to the new path
	if suffix != "" {
		newPath = newPath + "." + suffix
	}

	return newPath
}

package main

import (
"fmt"
"os"
"os/exec"
)

func displayImage(imgPath string) error {
	lowres := imgPath+".lowres.jpg"
	if !fileExists(lowres) {
		fmt.Println("caching! ", lowres)
		cacheImageInplace(imgPath)
	} 
	cmd := exec.Command("chafa", lowres)
	// cmd := exec.Command("ls", imgPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err) // Print error details for debugging
	}
	return err
}

func displayImage2(img Image) error {
	lowres := img.Lowr
	if !fileExists(lowres) {
		fmt.Println("caching! ", lowres)
		cacheImageInplace2(img)	
	}	 
	cmd := exec.Command("chafa", lowres)
	// cmd := exec.Command("ls", imgPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err) // Print error details for debugging
	}
	return err
}
func displayDumb(img Image) error {
	return nil
}

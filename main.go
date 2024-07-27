package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"flag"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
)

var (
	// Version variable to hold the version information
	Version = "0.1.8"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: gocuuull collecion.toml")
	}
	versionFlag := flag.Bool("v", false, "Show version")
	flag.Parse()
	
	if *versionFlag {
		fmt.Printf("Version: %s\n", Version)
		return
	}

	// CONFIG
	configFilePath := os.Args[1]
	config := loadConfig(configFilePath)

	imgs, err := listImages3(config)	
	if err != nil {
		log.Fatalf("Failed to list images: %v", err)
	}
	//log.Println(imgs)
	
	// images, err := listImages(config.LibraryPath, config.Users, fromDate, toDate)	
	// if err != nil {
	// 	log.Fatalf("Failed to list images: %v", err)
	// }

	// Sort images chronologically
	// sort.Strings(images)
	//log.Printf("Images found: %v", images)

	// Cache low-resolution images in the background
//	go cacheLowResImages(images)
	// log.Println("caching")
	go cacheLowResImages2(imgs)

	// Handle program interruption
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)

	if err := keyboard.Open(); err != nil {
		log.Fatalf("Failed to open keyboard: %v", err)
	}
	defer keyboard.Close()

	imageIndex := 0
	for {
		if imageIndex >= len(imgs) {
			fmt.Println("No more images to process.")
			return
		}

		// Display the current image
		if err := displayImage2(imgs[imageIndex]); err != nil {
		// if err := displayDumb(imgs[imageIndex]); err != nil {
			log.Printf("Failed to display image: %v", err)
			// Optionally continue or break depending on how you want to handle display errors
			imageIndex++
			continue
		}

		// Show image information
		printImageInfo2(imgs[imageIndex])

		select {
		case <-signalChannel:
			fmt.Println("\nQuitting program...")
			return
		default:
			char, key, err := keyboard.GetKey()
			if err != nil {
				log.Printf("Failed to get key: %v", err)
				continue
			}

			switch char {
			case 'k':
				if err := handleKeep2(imgs[imageIndex], config); err != nil {
					log.Printf("Failed to handle keep: %v", err)
				}
				imageIndex++ // Move to the next image
			case 'd':
				if err := handleDiscard2(imgs[imageIndex], config); err != nil {
					log.Printf("Failed to handle discard: %v", err)
				}
				imageIndex++ // Move to the next image
			case 'b':
				if imageIndex > 0 {
					imageIndex-- // Move back one image
				} else {
					fmt.Println("No previous images")
				}
			case 'q':
				fmt.Println("Quitting program...")
				return
			default:
				fmt.Println("Invalid input, please press 'k', 'd', or 'q'")
			}

			if key == keyboard.KeyArrowLeft {
				if imageIndex > 0 {
					imageIndex-- // Move back one image
				} else {
					fmt.Println("No previous images")
				}
			}

			// Ensure imageIndex is within bounds
			if imageIndex < 0 {
				imageIndex = 0
			}
		}
	}
}

func printImageInfo2(img Image) {
	fmt.Printf("\nFile: %s\nUser: %s\nDate: %s\n",
		filepath.Base(img.Path),
		img.User,
		img.Date.Format("2006-01-02 15:04:05"))
}


func printImageInfo(imgPath string) {
	pathParts := strings.Split(filepath.ToSlash(imgPath), "/")
	user := pathParts[len(pathParts)-5]
	dateStr := filepath.Join(pathParts[len(pathParts)-4], pathParts[len(pathParts)-3], pathParts[len(pathParts)-2])
	date, _ := time.Parse("2006/01/02", dateStr)

	fmt.Printf("\nFile: %s\nUser: %s\nDate: %s\n",
		filepath.Base(imgPath),
		user,
		date.Format("2006-01-02"))
}

func handleKeep2(img Image, config Config) error {
	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(img.Keep), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory tree: %v", err)
	}
	if err := touch(img.Seen); err != nil {
			return fmt.Errorf("failed to mark image as seen: %v", err)
	}	

	// if _, err := os.Stat(img.Keep); os.IsNotExist(err) {
	// 	return createHardLink(img.Path,img.Keep)
	// } else {
	
	if _, err := os.Stat(img.Keep); os.IsNotExist(err) {
		// Create the first hard link
	    if err := createHardLink(img.Path, img.Keep); err != nil {
	    	return err
	    }
	        // Create the second hard link
	        allLinkPath := filepath.Join(config.All, filepath.Base(img.Keep))
	        if err := createHardLink(img.Keep, allLinkPath); err != nil {
	            return err
	        }
	    } else {
	        fmt.Printf("Hard link already exists for: %s\n", filepath.Base(img.Path))
	    }
	
	    return nil
	
}
func handleDiscard2(img Image, config Config) error {
	if err := touch(img.Seen); err != nil {
			return fmt.Errorf("failed to mark image as seen: %v", err)
	}	
	// if _, err := os.Stat(img.Keep); !os.IsNotExist(err) {
	// 	return removeHardLink(img.Keep)
	// }

	if _, err := os.Stat(img.Keep); !os.IsNotExist(err) {
		// Remove the image
	    if err := removeHardLink(img.Keep); err != nil {
	    	return err
	    }
	    // Remove the image in the all folder
	        allLinkPath := filepath.Join(config.All, filepath.Base(img.Keep))
	        if err := removeHardLink(allLinkPath); err != nil {
	            return err
	        }
	    } 
	
	    return nil

}

func handleKeep(imgPath string, config Config) error {
	if err := touch(imgPath + ".seen"); err != nil {
			return fmt.Errorf("failed to mark image as seen: %v", err)
	}
	pathParts := strings.Split(filepath.ToSlash(imgPath), "/")
	user := pathParts[len(pathParts)-5]
	dateStr := filepath.Join(pathParts[len(pathParts)-4], pathParts[len(pathParts)-3], pathParts[len(pathParts)-2])

	// dst := filepath.Join(".", config.Collection, "kept", user, dateStr, filepath.Base(imgPath))
	dst := filepath.Join(".", config.Collection, user, dateStr, filepath.Base(imgPath))

	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return createHardLink(imgPath, dst)
	} else {
		fmt.Printf("Hard link already exists for: %s\n", filepath.Base(imgPath))
		return nil
	}
}

func handleDiscard(imgPath string, config Config) error {
	if err := touch(imgPath + ".seen"); err != nil {
			return fmt.Errorf("failed to mark image as seen: %v", err)
	}
	pathParts := strings.Split(filepath.ToSlash(imgPath), "/")
	user := pathParts[len(pathParts)-5]
	dateStr := filepath.Join(pathParts[len(pathParts)-4], pathParts[len(pathParts)-3], pathParts[len(pathParts)-2])

	// dst := filepath.Join(".", config.Collection, "kept", user, dateStr, filepath.Base(imgPath))
	dst := filepath.Join(".", config.Collection, user, dateStr, filepath.Base(imgPath))
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		return removeHardLink(dst)
	}
	return nil
}

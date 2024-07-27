package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"os/exec"

	// "github.com/dsoprea/go-exif/v3"
)

// func listImages3(libPath string, users []string, fromDate, toDate time.Time, collection, cache string) ([]Image, error) {
func listImages3(config Config) ([]Image, error) {
	var images []Image

	fmt.Println("Loading images...")
	
	libPath := config.LibraryPath
	users   := config.Users
	fromDate:= config.From
	toDate  := config.To
	collection := config.Collection
	cache   := config.Cache 

	err := filepath.Walk(libPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !isImageFile(info.Name()) {
			return nil
		}

		pathParts := strings.Split(filepath.ToSlash(path), "/")
		if len(pathParts) < 5 {
			return nil
		}

		user := pathParts[len(pathParts)-5]
		if !contains(users, user) {
			return nil
		}

		
		dateStr := filepath.Join(pathParts[len(pathParts)-4], pathParts[len(pathParts)-3], pathParts[len(pathParts)-2])
		date, err := time.Parse("2006/01/02", dateStr)
		if err != nil {
			return nil
		}


		if !date.Before(fromDate) && !date.After(toDate) {
		


			img := Image{
				Path:    path,
				User:    user,
				Date:    date,
				// Creation: creationDate,
				Keep:    buildPath(path, libPath, collection, ""),
				Lowr:    buildPath(path, libPath, cache, "lowres.jpg"),
				Seen:    buildPath(path, libPath, cache, "seen"),
			}

			if !fileExists(img.Seen) {
				creationDate, err := extractDateFromExif(path)
				if err != nil {
					fmt.Println("Error extracting EXIF date:", err)
					return nil
				}
				img.Creation = creationDate
			}

			if !fileExists(img.Seen) {
				images = append(images, img)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	fmt.Println("Sorting ",len(images), "images")
	sort.Slice(images, func(i, j int) bool {
		return images[i].Creation.Before(images[j].Creation)
	})

	// for _, img := range images {
	// 	fmt.Println(img.Creation, img.Path)
	// }

	return images, nil
}


func extractDateFromExif(path string) (time.Time, error) {
	// Run exiftool command to extract DateTimeOriginal
	cmd := exec.Command("exiftool", "-DateTimeOriginal", path)

    // Run the command and capture the output
    output, err := cmd.Output()
    if err != nil {
        fmt.Println("Error executing command:", err)
        // return
    }
    
    // Convert the output to a string
    outputStr := string(output)
    // fmt.Println("Raw Output:", outputStr)

    // Extract the date/time part from the output
    var dateTimeStr string
    lines := strings.Split(outputStr, "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "Date/Time Original") {
            // Extract the date/time part (everything after the colon and any following spaces)
            parts := strings.SplitN(line, ":", 2)
            if len(parts) > 1 {
                dateTimeStr = strings.TrimSpace(parts[1])
            }
            break
        }
    }

    // If we didn't find the date/time, handle the error
    if dateTimeStr == "" {
        fmt.Println("Date/Time not found in output")
        // return
    }

    // Define the layout format to match the extracted date/time string
    layout := "2006:01:02 15:04:05"

    // Parse the string into a time.Time object
    t, err := time.Parse(layout, dateTimeStr)
    if err != nil {
        fmt.Println("Error parsing date:", err)
        // return
    }

	return t, nil
}

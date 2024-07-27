package main

import (
	"log"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"io"
	"syscall"
)

type Image struct {
	Path string
	User string
	Date time.Time
	Keep string
	// placeholders
	Seen string
	Lowr string
	Creation time.Time

}

var imageExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".gif":  {},
	".bmp":  {},
	".tiff": {},
	".tif":  {},
	".heic": {},
	".webp": {},
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, isImage := imageExtensions[ext]
	return isImage
}

func listImages2(config Config) ([]Image, error) {
	var images []Image
	var img      Image
	
	libPath := config.LibraryPath
	err := filepath.Walk(libPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if the path matches the user directory structure
		pathParts := strings.Split(filepath.ToSlash(path), "/")
		if len(pathParts) < 5 {
			return nil
		}

		// if strings.HasSuffix(info.Name(), ".xmp") {
		// 	return nil
		// }
		// if strings.HasSuffix(info.Name(), ".mov") {
		// 	return nil
		// }
		// if strings.HasSuffix(info.Name(), ".mp4") {
		// 	return nil
		// }
		
		if !isImageFile(info.Name()) {
				return nil
		}

		// WARNING!----------------------------------------
		// DONT NEED THIS since lowresolution and seen files 
		// are stored in the cache folder
		//-------------------------------------------------
		// // Exclude low-resolution files
		// if strings.HasSuffix(info.Name(), ".lowres.jpg") {
		// 	return nil
		// }
		// if strings.HasSuffix(info.Name(), ".seen") {
		// 	return nil
		// }

		// USERS---------------------------
		users := config.Users
		user := pathParts[len(pathParts)-5]
		if !contains(users, user) {
			return nil
		}

		// DATES--------------------------
		fromDate := config.From
		toDate   := config.To

		dateStr := filepath.Join(pathParts[len(pathParts)-4], pathParts[len(pathParts)-3], pathParts[len(pathParts)-2])
		date, err := time.Parse("2006/01/02", dateStr)
		if err != nil {
			return nil
		}

		// Check if the date is within the range
		if !date.Before(fromDate) && !date.After(toDate) {
				img.Path = path
				img.User = user
				img.Date = date
				img.Keep = buildPath(path,config.LibraryPath,config.Collection, "") 
				img.Lowr = buildPath(path,config.LibraryPath,config.Cache, "lowres.jpg") 
				img.Seen = buildPath(path,config.LibraryPath,config.Cache, "seen") 
			if !fileExists(img.Seen) {
				images = append(images, img)
			}
		}
		return nil
	})
	return images, err
}





func listImages(libPath string, users []string, fromDate, toDate time.Time) ([]string, error) {
	var images []string
	err := filepath.Walk(libPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if the path matches the user directory structure
		pathParts := strings.Split(filepath.ToSlash(path), "/")
		if len(pathParts) < 5 {
			return nil
		}

		// Exclude low-resolution files
		if strings.HasSuffix(info.Name(), ".lowres.jpg") {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".seen") {
			return nil
		}

		// Extract date from the path
		user := pathParts[len(pathParts)-5]
		if !contains(users, user) {
			return nil
		}

		dateStr := filepath.Join(pathParts[len(pathParts)-4], pathParts[len(pathParts)-3], pathParts[len(pathParts)-2])
		date, err := time.Parse("2006/01/02", dateStr)
		if err != nil {
			return nil
		}

		// Check if the date is within the range
		if !date.Before(fromDate) && !date.After(toDate) {
			if !fileExists(path + ".seen") {
				images = append(images, path)
			}
		}
		return nil
	})
	return images, err
}

func cacheLowResImages2(images []Image) {
	for _, img := range images {
		cacheImageInplace2(img)
	}
}

func cacheImageInplace2(img Image) {
			lowResPath := img.Lowr
			if err := os.MkdirAll(filepath.Dir(lowResPath), os.ModePerm); err != nil {
			    log.Fatalf("failed to create cache tree structure: %v", err)
			}
			if _, err := os.Stat(lowResPath); os.IsNotExist(err) {
				cmd := exec.Command("magick", img.Path, "-resize", "800x800", lowResPath)
				if err := cmd.Run(); err != nil {
					log.Println("Failed to cache low-resolution image: " + img.Path)
					log.Println(err)
				}
			}		
}

func cacheLowResImages(images []string) {
	for _, img := range images {
		cacheImageInplace(img)
	}
}

func cacheImageInplace(img string) {
	lowResPath := img + ".lowres.jpg"
	if _, err := os.Stat(lowResPath); os.IsNotExist(err) {
		cmd := exec.Command("magick", img, "-resize", "800x800", lowResPath)
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to cache low-resolution image: %v", err)
		}
	}		
}

// func createHardLink(src, dst string) error {
// 	return os.Link(src, dst)
// }

// createHardLink creates a hard link for the source file if on the same filesystem, otherwise copies the file
func createHardLink(src, dst string) error {
	// Check if source and target are on the same filesystem
	sameFs, err := checkSameFilesystem(src, filepath.Dir(dst))
	if err != nil {
		return fmt.Errorf("error checking filesystem: %v", err)
	}

	if sameFs {
		// Create a hard link
		return os.Link(src, dst)
	} else {
		// Copy the file
		return copyFile(src, dst)
	}

}

// checkSameFilesystem checks if two paths are on the same filesystem
func checkSameFilesystem(src, dstDir string) (bool, error) {
	var srcStat, dstStat syscall.Stat_t

	if err := syscall.Stat(src, &srcStat); err != nil {
		return false, err
	}
	if err := syscall.Stat(dstDir, &dstStat); err != nil {
		return false, err
	}

	return srcStat.Dev == dstStat.Dev, nil
}

// copyFile copies the source file to the target location
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return dstFile.Sync()
}

func removeHardLink(dst string) error {
	return os.Remove(dst)
}

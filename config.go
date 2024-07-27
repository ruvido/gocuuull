package main

import (
	"github.com/BurntSushi/toml"
	"log"
	// "fmt"
	"time"
	"os"
	// "path/filepath"
)

type Config struct {
	LibraryPath string   `toml:"lib"`
	FromDate 	string   `toml:"from"`
	ToDate   	string   `toml:"to"`
	Collection  string   `toml:"name"`
	Users       []string `toml:"users"`
	LowSuffix	string 	 `toml:"lowsuffix"`
	SeeSuffix	string   `toml:"seesuffix"`
	Cache 		string   `toml:"cache"`
	From		time.Time
	To			time.Time
	All 		string   `toml:"all"` // this is a local folder inside Collection where all selected images are hard linked
}

func loadConfig(path string) Config {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	if config.Cache == "" {
		config.Cache = config.Collection+"/.cache"
	}
	if err := os.MkdirAll(config.Cache, os.ModePerm); err != nil {
		log.Fatalf("failed to create cache directory: %v", err)
		os.Exit(99)
	}
	
	if config.All == "" {
		config.All = config.Collection+"/all"
	}
	if err := os.MkdirAll(config.All, os.ModePerm); err != nil {
		log.Fatalf("failed to create all saved images directory: %v", err)
		os.Exit(99)
	}

	if config.LowSuffix == "" {
		config.LowSuffix = ".lowres.jpg"
	}
	if config.SeeSuffix == "" {
		config.SeeSuffix = ".seen"
	}

	fromDate, err := time.Parse("20060102", config.FromDate)
	if err != nil {
		log.Fatalf("Invalid from date format: %v", err)
	}
	config.From = fromDate
	
	toDate, err := time.Parse("20060102", config.ToDate)
	if err != nil {
		log.Fatalf("Invalid to date format: %v", err)
	}
	config.To = toDate

	// create directory structure
	
	return config
}

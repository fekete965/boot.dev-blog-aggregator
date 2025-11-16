package main

import (
	"fmt"
	"log"

	"github.com/fekete965/boot.dev-blog-aggregator/internal/config"
)

func main() {
	configFile, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	
	configFile.SetUser("Bence")

	configFile, err = config.Read()
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	fmt.Printf("Config: %+v\n", configFile)
}

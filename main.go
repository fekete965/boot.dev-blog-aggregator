package main

import (
	"fmt"
	"log"

	"github.com/fekete965/boot.dev-blog-aggregator/internal/config"
)

type state struct {
	Config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(s *state, cmd command) error
}

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

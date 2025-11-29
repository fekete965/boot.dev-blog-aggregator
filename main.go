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

func (c *commands) run(s *state, cmd command) error {
	if fn, ok := c.handlers[cmd.name]; ok {
		return fn(s, cmd)
	}

	return fmt.Errorf("command not found: %v", cmd.name)
}

func (c *commands) register(name string, fn func(s *state, cmd command) error) error {
	if _, ok := c.handlers[name]; ok {
		return fmt.Errorf("command already registered: %v", name)
	}

	c.handlers[name] = fn
	return nil
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

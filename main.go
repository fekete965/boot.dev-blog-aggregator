package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/fekete965/boot.dev-blog-aggregator/internal/config"
	"github.com/fekete965/boot.dev-blog-aggregator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	config *config.Config
	database *database.Queries 
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

func handleLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("the login command requires a username. Usage: gator login <username>")
	}

	username := cmd.args[0]

	existingUser, err := s.database.FindUserByeName(context.Background(), username)
	if err != nil {
		log.Fatalf("cannot find user: %v", err)
	}

	if err := s.config.SetUser(existingUser.Name); err != nil {
		return fmt.Errorf("an error occurred during login: %v", err)
	}

	fmt.Printf("Current user has been set to: %v\n", existingUser.Name)
	return nil
}
	return nil
}

func main() {
	configFile, err := config.Read()
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	// Create the database connection
	db, err := sql.Open("postgres", configFile.DBUrl)
	if err != nil {
		log.Fatalf("failed to create database connection: %v", err)
	}
	dbQueries := database.New(db)

	// Initialize the application state
	appState := &state {
		database: dbQueries, 
		config: configFile,
	}

	// Initialize the command handlers
	commands := &commands {
		handlers: make(map[string]func(s *state, cmd command) error),
	}

	// Register the command handlers
	commands.register("login", handleLogin)

	if len(os.Args) < 2 {
		log.Fatalf("you did not provide any arguments. Usage of gator is: gator <command> <args>")
	}

	newCommand := command {
		name: os.Args[1],
		args: os.Args[2:],
	}

	if err := commands.run(appState, newCommand); err != nil {
		log.Fatalf("failed to run command: %v: ", err)
	}

	configFile, err = config.Read()
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	fmt.Printf("Config: %+v\n", configFile)
}

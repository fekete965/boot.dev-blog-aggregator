package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fekete965/boot.dev-blog-aggregator/internal/config"
	"github.com/fekete965/boot.dev-blog-aggregator/internal/database"
	"github.com/google/uuid"

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

func handleRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("the register command requires a username. Usage: gator register <username>")	
	}

	username := cmd.args[0]

	createdUser, err := s.database.CreateUser(context.Background(), database.CreateUserParams{
		ID: uuid.New(),
		Name: username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		log.Fatalf("failed to register user: %v", err)
	}

	if err := s.config.SetUser(createdUser.Name); err != nil {
		return fmt.Errorf("an error occurred during registration: %v", err)
	}

	fmt.Printf("User has been registered: %v\n", username)
	fmt.Printf("User ID: %v\n", createdUser.ID)
	fmt.Printf("User Name: %v\n", createdUser.Name)
	fmt.Printf("User Created At: %v\n", createdUser.CreatedAt)
	fmt.Printf("User Updated At: %v\n", createdUser.UpdatedAt)

	return nil
}

func handleReset(s *state, cmd command) error {
	err := s.database.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to reset database: %v", err)
	}

	fmt.Println("Database has been reset")

	return nil
}

func handleUsers(s *state, cmd command) error {
	users, err := s.database.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get users: %v", err)
	}

	for _, user := range users {
		outputName := user.Name
		
		if *s.config.CurrentUserName == user.Name {
			outputName += " (current)"
		}
		
		fmt.Printf("* %v\n", outputName)
	}

	return nil
}

func handleAggregate(s *state, cmd command) error {
	// For the future:
	// if s.config.CurrentUserName == nil || *s.config.CurrentUserName == "" {
	// 	return fmt.Errorf("you must be logged in to aggregate feeds. Usage: gator login <username> or gator register <username>")
	// }
	
	// if len(cmd.args) == 0 {
	// 	return fmt.Errorf("the aggregate command requires feed URL. Usage: gator agg <feed_url>")
	// }

	// feedUrl := cmd.args[0]
	feedUrl := "https://www.wagslane.dev/index.xml"

	rssFeed, err := fetchFeed(context.Background(), feedUrl)
	if err != nil {
		return fmt.Errorf("failed to aggregate feed: %v", err)
	}

	fmt.Printf("Feed successfully aggregated\n")
	fmt.Printf("%-12s %v\n", "Title:", html.UnescapeString(rssFeed.Channel.Title))
	fmt.Printf("%-12s %v\n", "Description:", html.UnescapeString(rssFeed.Channel.Description))
	fmt.Printf("%-12s %v\n", "Link:", html.UnescapeString(rssFeed.Channel.Link))
	
	fmt.Printf("Items \n")
	for i, item := range rssFeed.Channel.Item {
		fmt.Printf(" - %v. %-12s %v\n", i + 1, "Title:", html.UnescapeString(item.Title))
		fmt.Printf(" -     %-12s %v\n", "Description:", html.UnescapeString(item.Description))
		fmt.Printf(" -     %-12s %v\n", "Link:", html.UnescapeString(item.Link))
		fmt.Printf(" -     %-12s %v\n", "PubDate:", html.UnescapeString(item.PubDate))
	}

	return nil
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Item        []RSSItem `xml:"item"`
}

type RSSFeed struct {
	Channel RSSChannel `xml:"channel"`
}

func fetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("something went wrong creating the request: %v", err)
	}

	req.Header.Set("User-Agent", "gator/1.0")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("something went wrong fetching the feed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("something went wrong reading the response body: %v", err)
	}
	
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("response failed with status code: %v and body %v", resp.StatusCode, body)
	}

	var result RSSFeed = RSSFeed{}
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("body marshalling failed: %v", err)
	}

	return &result, nil
}

func handleAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("the addfeed command requires a name and feed URL. Usage: gator addfeed <feed_name> <feed_url>")
	}

	feedName := cmd.args[0]
	feedUrl := cmd.args[1]
	
	currentUser, err := s.database.FindUserByeName(context.Background(), *s.config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failed to find current user: %v", err)
	}

	createdFeed, err := s.database.CreateFeed(context.Background(), database.CreateFeedParams{
		ID: uuid.New(),
		UserID: currentUser.ID,
		Name: feedName,
		Url: feedUrl,
	})
	if err != nil {
		return fmt.Errorf("failed to create feed: %v", err)
	}

	fmt.Printf("Feed successfully added\n")
	fmt.Printf("- Id: %v\n", createdFeed.ID)
	fmt.Printf("- User ID: %v\n", createdFeed.UserID)
	fmt.Printf("- Name: %v\n", createdFeed.Name)
	fmt.Printf("- URL: %v\n", createdFeed.Url)

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
	commands.register("register", handleRegister)
	commands.register("reset", handleReset)
	commands.register("users", handleUsers)
	commands.register("agg", handleAggregate)
	commands.register("addfeed", handleAddFeed)

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
}

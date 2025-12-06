package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/fekete965/boot.dev-blog-aggregator/internal/config"
	"github.com/fekete965/boot.dev-blog-aggregator/internal/database"
	"github.com/google/uuid"

	"github.com/lib/pq"
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

var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")
var ErrNoNextFeedFound = errors.New("no next feed found")
var ErrPostExists = errors.New("post already exists")

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

func loginUser(s *state, username string) (database.User, error) {
	user, err := s.database.FindUserByeName(context.Background(), username)
	if err != nil {
		if (errors.Is(err, sql.ErrNoRows)) {
			return database.User{}, ErrUserNotFound
		}

		return database.User{}, fmt.Errorf("failed to login: %v", err)
	}

	return user, nil
}

func handleLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("the login command requires a username. Usage: gator login <username>")
	}

	username := cmd.args[0]

	existingUser, err := loginUser(s, username)
	if err != nil {
		if err == ErrUserNotFound {
			return fmt.Errorf("user not found. Please double check the username and try again")
		}

		return fmt.Errorf("failed to login: %v", err)
	}

	if err := s.config.SetUser(existingUser.Name); err != nil {
		return fmt.Errorf("an error occurred during login: %v", err)
	}

	fmt.Printf("Current user has been set to: %v\n", existingUser.Name)
	return nil
}

func registerUser(s *state, username string) (database.User, error) {
	newUser, err := s.database.CreateUser(context.Background(), database.CreateUserParams{
		ID: uuid.New(),
		Name: username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return database.User{}, ErrUserAlreadyExists
		}

		return database.User{}, fmt.Errorf("failed to register user: %v", err)
	}

	return newUser, nil
}

func handleRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("the register command requires a username. Usage: gator register <username>")	
	}

	username := cmd.args[0]

	createdUser, err := registerUser(s, username)
	if err != nil {
		if err == ErrUserAlreadyExists {
			return fmt.Errorf("user already exists. Please use a different username")
		}

		return fmt.Errorf("failed to register user: %v", err)
	}

	if err := s.config.SetUser(createdUser.Name); err != nil {
		return fmt.Errorf("an error occurred during registration: %v", err)
	}

	fmt.Printf("New user registered: %v\n", username)

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
	if len(cmd.args) == 0 {
		return fmt.Errorf("the aggregate command requires time between requests.\nValid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\".\nUsage: gator agg <time_between_requests>")
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("failed to parse time between requests: %v", err)
	}

	fmt.Printf("Collecting feeds every %v\n", timeBetweenReqs)

	ticker := time.NewTicker(timeBetweenReqs)

	for ; ; <- ticker.C {
		err := scrapeFeeds(s)
		if errors.Is(err, ErrNoNextFeedFound) {
			fmt.Println("No feeds to scrape")
			return nil
		}
		if err != nil {
			return fmt.Errorf("error scraping feeds: %v", err)
		}
	}
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

func handleAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("the addfeed command requires a name and feed URL. Usage: gator addfeed <feed_name> <feed_url>")
	}

	feedName := cmd.args[0]
	feedUrl := cmd.args[1]

	createdFeed, err := s.database.CreateFeed(context.Background(), database.CreateFeedParams{
		ID: uuid.New(),
		UserID: user.ID,
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

	newFeedFollow, err := createFeedFollowForUser(s, createdFeed.ID, user.ID)
	if err != nil {
		return fmt.Errorf("failed to follow the feed: %v", err)
	}

	fmt.Printf("Successfully followed the feed: %v\n", newFeedFollow.FeedName)

	return nil
}

func handleFeeds(s *state, cmd command) error {
	feeds, err := s.database.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get feeds: %v", err)
	}

	fmt.Printf("Current feeds\n")
	fmt.Printf("--------------------------------\n")
	for _, feedData := range feeds {
		fmt.Printf("- Feed Id:   %v\n", feedData.FeedID)
		fmt.Printf("- User Name: %v\n", feedData.UserName)
		fmt.Printf("- Feed Name: %v\n", feedData.FeedName)
		fmt.Printf("- Feed URL:  %v\n", feedData.FeedUrl)
		fmt.Printf("--------------------------------\n")
	}

	return nil
}

func createFeedFollowForUser(s *state, feedID uuid.UUID, userID uuid.UUID) (*database.CreateFeedFollowRow, error) {

	newFeedFollow, err := s.database.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		UserID: userID,
		FeedID: feedID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create feed follow: %v", err)
	}

	return &newFeedFollow, nil
}

func handleFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("the follow requires a feed URL. Usage: gator follow <feed_url>")
	}

	feedUrl := cmd.args[0]

	feed, err := s.database.FindFeedByUrl(context.Background(), feedUrl)
	if err != nil {
		return fmt.Errorf("failed to get feed by URL: %v", err)
	}

	newFeedFollow, err := createFeedFollowForUser(s, feed.ID, user.ID)
	if err != nil {
		return fmt.Errorf("failed to follow the feed: %v", err)
	}

	fmt.Printf("Successfully followed the feed: %v\n", newFeedFollow.FeedName)

	return nil
}

func handleFollowing(s *state, cmd command, user database.User) error {
	feedFollows, err := s.database.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get feed follows for user: %v", err)
	}

	feedLength := len(feedFollows)
	postfix := "feed"
	if feedLength != 1 {
		postfix += "s"
	}
	fmt.Printf("Current user is following %v %v\n", feedLength, postfix)

	for _, feedFollow := range feedFollows {
		fmt.Printf("- %v", feedFollow.FeedName)
	}

	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(s *state, cmd command) error {
	var responseFn = func(s *state, cmd command) error {
			if s.config.CurrentUserName == nil || *s.config.CurrentUserName == "" {
				return fmt.Errorf("you must be logged in to use this feature. Use: gator login <username> or gator register <username>")
			}
			
			currentUser, err := s.database.FindUserByeName(context.Background(), *s.config.CurrentUserName)
			if err != nil {
				return fmt.Errorf("failed to find current user: %v", err)
			}

			return handler(s, cmd, currentUser)
	}

	return responseFn
}

func handleUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("the unfollow command required a feed URL. Usage gator unfollow <feed_url>")
	}

	feedUrl := cmd.args[0]
	feed, err := s.database.FindFeedByUrl(context.Background(), feedUrl)
	if err != nil {
		return fmt.Errorf("failed to find feed by URL: %v", err)
	}
	err = s.database.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to unfollow the feed: %v", err)
	}

	fmt.Printf("Successfully unfollowed the feed: %v\n", feed.Name)

	return nil
}

func parsePubDate(dateStr string) (time.Time, error) {
	formats := []string{
			time.RFC1123Z,
			time.RFC1123,
			time.RFC822Z,
			time.RFC822,
			time.RFC3339,
	}
	
	for _, format := range formats {
			if t, err := time.Parse(format, dateStr); err == nil {
					return t, nil
			}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func scrapeFeeds(s *state) error {
	nextFeed, err := s.database.GetNextFeedToFetch(context.Background())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoNextFeedFound
		}
		
		return fmt.Errorf("failed to get next feed: %v", err)
	}
	
	err = s.database.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		ID: nextFeed.ID,
		LastFetchedAt: sql.NullTime{
			Time: time.Now(),
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to mark feed fetched: %v", err)
	}

	rssFeed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %v", err)
	}
	
	for _, item := range rssFeed.Channel.Item {
		publishedAt, dateParsingErr := parsePubDate(item.PubDate)

		err = createPost(s, database.CreatePostParams{
			ID: uuid.New(),
			FeedID: nextFeed.ID,
			Title: html.UnescapeString(item.Title),
			Url: html.UnescapeString(item.Link),
			Description: html.UnescapeString(item.Description),
			PublishedAt: sql.NullTime{
				Time: publishedAt,
				Valid: dateParsingErr == nil,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})

		if err == ErrPostExists {
			fmt.Printf("post already exists: %v\n", item.Title)
			continue
		}

		if err != nil {
			return fmt.Errorf("failed to scrape post: %v", err)
		}
	}

	return nil
}

func createPost(s *state, data database.CreatePostParams) error {
	newPost, err := s.database.CreatePost(context.Background(), data)
	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrPostExists
		}

		return fmt.Errorf("failed to create post: %v", err)
	}

	fmt.Printf("Post successfully created: %v\n", newPost.Title)

	return nil
}

func truncateString(str string, maxLength int) string {
	if len(str) <= maxLength {
		return str
	}

	return str[:maxLength - 3] + "..."
}

func handleBrowse(s *state, cmd command, user database.User) error {
	postLimitArg := "2"
	if len(cmd.args) > 0 {
		postLimitArg = cmd.args[0]
	}

	postLimit, err := strconv.Atoi(postLimitArg)
	if err != nil {
		fmt.Printf("failed to convert post limit to int: %v\n", err)
		postLimit = 2
	}

	userPosts, err := s.database.GetPostForUser(context.Background(), database.GetPostForUserParams{
		UserID: user.ID,
		Limit: int32(postLimit),
	})
	if err != nil {
		return fmt.Errorf("failed to get posts for the user: %v", err)
	}

	if len(userPosts) == 0 {
		fmt.Println("No posts found")
		return nil
	}

	for _, post := range userPosts {
		publishedAtStr := "N/A"
		if post.PublishedAt.Valid {
			publishedAtStr = post.PublishedAt.Time.Format("02 January 2006 15:04")
		}

		fmt.Println("┌─────────────────────────────────────────────────────────────┐")
		fmt.Printf("│ Title: %-50s │\n", truncateString(post.Title, 50))
		fmt.Printf("│ Published At: %-44s │\n", publishedAtStr)
		fmt.Println("├─────────────────────────────────────────────────────────────┤")
		fmt.Printf("│ Description: %-51s │\n", "")
		fmt.Printf("│ %-59s │\n", truncateString(post.Description, 59))
		fmt.Println("├─────────────────────────────────────────────────────────────┤")
		fmt.Printf("│ URL: %-55s │\n", truncateString(post.Url, 55))
		fmt.Println("└─────────────────────────────────────────────────────────────┘")
		fmt.Println()
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
	commands.register("register", handleRegister)
	commands.register("reset", handleReset)
	commands.register("users", handleUsers)
	commands.register("agg", handleAggregate)
	commands.register("addfeed", middlewareLoggedIn(handleAddFeed))
	commands.register("feeds", handleFeeds)
	commands.register("follow", middlewareLoggedIn(handleFollow))
	commands.register("following", middlewareLoggedIn(handleFollowing))
	commands.register("unfollow", middlewareLoggedIn(handleUnfollow))
	commands.register("browse", middlewareLoggedIn(handleBrowse))
	
	if len(os.Args) < 2 {
		log.Fatalf("you did not provide any arguments. Usage of gator is: gator <command> <args>")
	}

	newCommand := command {
		name: os.Args[1],
		args: os.Args[2:],
	}

	if err := commands.run(appState, newCommand); err != nil {
		log.Fatalf("failed to run command: \"%v\"\n%v", newCommand.name, err)
	}
}

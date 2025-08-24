package gameFetch

import (
	"bufio"
	"chessAnalyserFree/api"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// GameFetcher manages the state of fetching and displaying games.
type GameFetcher struct {
	client         *api.Client
	username       string
	allGames       []api.Game
	displayedCount int
	currentDate    time.Time
}

// NewGameFetcher creates a new instance of GameFetcher.
func NewGameFetcher(client *api.Client, username string) *GameFetcher {
	return &GameFetcher{
		client:         client,
		username:       username,
		allGames:       []api.Game{},
		displayedCount: 0,
		currentDate:    time.Now(),
	}
}

// fetchGamesForMonth fetches games for the month specified by the fetcher's currentDate.
func (f *GameFetcher) fetchGamesForMonth() error {
	year := f.currentDate.Format("2006")
	month := f.currentDate.Format("01")

	fmt.Printf("Fetching games for %s/%s...\n", month, year)
	monthlyGames, err := f.client.FetchPlayerGamesByMonth(f.username, year, month)
	if err != nil {
		// Don't treat a 404 as a fatal error, it just means no games for that month.
		if strings.Contains(err.Error(), "status code: 404") {
			fmt.Printf("No games found for %s in %s/%s.\n", f.username, month, year)
		} else {
			return fmt.Errorf("error fetching games for %s/%s: %w", month, year, err)
		}
	}

	if monthlyGames != nil {
		// The API returns games in chronological order, so we reverse to get the latest first.
		for i := len(monthlyGames.Games) - 1; i >= 0; i-- {
			f.allGames = append(f.allGames, monthlyGames.Games[i])
		}
	}

	// Move to the previous month for the next fetch operation.
	f.currentDate = f.currentDate.AddDate(0, -1, 0)
	return nil
}

// ensureEnoughGames makes sure we have at least a certain number of games fetched.
func (f *GameFetcher) ensureEnoughGames(count int) error {
	for len(f.allGames) < count {
		// Stop if we go back more than 10 years, to prevent infinite loops for inactive users.
		if f.currentDate.Before(time.Now().AddDate(-10, 0, 0)) {
			fmt.Println("Searched 10 years of history. No more games found.")
			return nil
		}
		if err := f.fetchGamesForMonth(); err != nil {
			return err
		}
	}
	return nil
}

// displayGames shows a slice of the fetched games.
func (f *GameFetcher) displayGames() {
	if f.displayedCount >= len(f.allGames) {
		fmt.Println("No more games to display.")
		return
	}

	fmt.Println("\n--- Latest Games ---")
	limit := f.displayedCount + 10
	if limit > len(f.allGames) {
		limit = len(f.allGames)
	}

	for i := f.displayedCount; i < limit; i++ {
		game := f.allGames[i]
		endTime := time.Unix(game.EndTime, 0)
		fmt.Printf("[%d] %s vs %s (%s) - Played on %s\n",
			i+1, game.White.Username, game.Black.Username, game.TimeClass, endTime.Format("2006-01-02"))
	}
	f.displayedCount = limit
	fmt.Println("--------------------")
}

// displayGameDetails shows detailed information for a selected game.
func (f *GameFetcher) displayGameDetails(index int) {
	if index < 0 || index >= len(f.allGames) {
		fmt.Println("Invalid game number.")
		return
	}

	game := f.allGames[index]
	endTime := time.Unix(game.EndTime, 0)

	fmt.Printf("\n--- Game Details (%d) ---\n", index+1)
	fmt.Printf("URL: %s\n", game.URL)
	fmt.Printf("Date: %s\n", endTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Time Class: %s (%s)\n", game.TimeClass, game.TimeControl)
	fmt.Printf("Rules: %s\n", game.Rules)
	fmt.Printf("Rated: %t\n", game.Rated)
	fmt.Printf("White: %s (%d) - Result: %s\n", game.White.Username, game.White.Rating, game.White.Result)
	fmt.Printf("Black: %s (%d) - Result: %s\n", game.Black.Username, game.Black.Rating, game.Black.Result)
	fmt.Printf("Final Position (FEN): %s\n", game.FEN)
	fmt.Println("--- PGN ---")
	fmt.Println(game.PGN)
	fmt.Println("------------------------")
}

// Run starts the interactive command-line interface.
func (f *GameFetcher) Run() {
	// Fetch initial batch of games
	if err := f.ensureEnoughGames(10); err != nil {
		fmt.Printf("Failed to fetch initial games: %v\n", err)
		return
	}
	f.displayGames()

	// Start interactive loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command ('more', 'select [number]', or 'quit'): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		command := strings.ToLower(parts[0])

		switch command {
		case "quit":
			fmt.Println("Goodbye!")
			return
		case "more":
			if err := f.ensureEnoughGames(f.displayedCount + 10); err != nil {
				fmt.Printf("Failed to fetch more games: %v\n", err)
			}
			f.displayGames()
		case "select":
			if len(parts) < 2 {
				fmt.Println("Please provide a game number to select.")
				continue
			}
			gameNum, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Invalid game number provided.")
				continue
			}
			f.displayGameDetails(gameNum - 1) // Adjust for 0-based index
		default:
			fmt.Println("Unknown command. Available commands: 'more', 'select [number]', 'quit'.")
		}
	}
}

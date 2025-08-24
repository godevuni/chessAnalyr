package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// baseURL is the base URL for the Chess.com public data API.
const baseURL = "https://api.chess.com/pub"

// Client is a client for the Chess.com API.
type Client struct {
	HTTPClient *http.Client
}

// NewClient creates a new Chess.com API client.
func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Player holds the details for the white or black player in a game.
type Player struct {
	Rating   int    `json:"rating"`
	Result   string `json:"result"`
	ID       string `json:"@id"`
	Username string `json:"username"`
}

// Game represents a single game played on Chess.com.
type Game struct {
	URL         string `json:"url"`
	PGN         string `json:"pgn"`
	TimeControl string `json:"time_control"`
	EndTime     int64  `json:"end_time"`
	Rated       bool   `json:"rated"`
	FEN         string `json:"fen"`
	TimeClass   string `json:"time_class"`
	Rules       string `json:"rules"`
	White       Player `json:"white"`
	Black       Player `json:"black"`
}

// GamesResponse is the structure of the JSON response for the monthly games archive.
type GamesResponse struct {
	Games []Game `json:"games"`
}

// FetchPlayerGamesByMonth fetches all games for a given player for a specific year and month.
// The year should be in YYYY format (e.g., "2022").
// The month should be in MM format (e.g., "01" for January).
func (c *Client) FetchPlayerGamesByMonth(username, year, month string) (*GamesResponse, error) {
	// Construct the request URL.
	url := fmt.Sprintf("%s/player/%s/games/%s/%s", baseURL, username, year, month)

	// Create a new HTTP request.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// It's good practice to set a User-Agent header.
	req.Header.Set("User-Agent", "Go-Chess.com-API-Client/1.0 (your-contact-info)")

	// Execute the request.
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a successful status code.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	// Read the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal the JSON response into our struct.
	var gamesResponse GamesResponse
	if err := json.Unmarshal(body, &gamesResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	return &gamesResponse, nil
}

// Example usage:
// func main() {
// 	client := NewClient()
// 	username := "hikaru" // Example username
// 	year := "2022"
// 	month := "12"
//
// 	games, err := client.FetchPlayerGamesByMonth(username, year, month)
// 	if err != nil {
// 		log.Fatalf("Error fetching games: %v", err)
// 	}
//
// 	fmt.Printf("Found %d games for %s in %s/%s\n", len(games.Games), username, month, year)
// 	for i, game := range games.Games {
// 		fmt.Printf("Game %d: %s vs %s - URL: %s\n", i+1, game.White.Username, game.Black.Username, game.URL)
// 	}
// }

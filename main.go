package main

import (
	"bufio"
	"chessAnalyserFree/api"
	gameengine "chessAnalyserFree/gameEngine"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	// --- Argument Parsing ---
	// Expected format: go run . <username> <start_YYYY-MM> <end_YYYY-MM> <path_to_stockfish>
	if len(os.Args) != 5 {
		fmt.Println("Usage: go run . <username> <start_YYYY-MM> <end_YYYY-MM> <path_to_stockfish>")
		fmt.Println("Example: go run . hikaru 2022-10 2023-01 /usr/local/bin/stockfish")
		return
	}

	username := os.Args[1]
	startDateStr := os.Args[2]
	endDateStr := os.Args[3]
	stockfishPath := os.Args[4]

	// --- Stockfish Analyser Initialization ---
	analyser, err := gameengine.NewStockfishAnalyser(stockfishPath)
	if err != nil {
		log.Fatalf("Error starting Stockfish analyser: %v", err)
	}
	defer analyser.Close()
	fmt.Println("Stockfish engine initialized successfully.")

	// --- Date Parsing ---
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, startDateStr+"-01")
	if err != nil {
		log.Fatalf("Error parsing start date: %v. Please use YYYY-MM format.", err)
	}
	endDate, err := time.Parse(layout, endDateStr+"-01")
	if err != nil {
		log.Fatalf("Error parsing end date: %v. Please use YYYY-MM format.", err)
	}

	if startDate.After(endDate) {
		log.Fatal("Start date cannot be after the end date.")
	}

	// --- API Client Initialization ---
	client := api.NewClient()
	var allGames []api.Game
	totalGamesFound := 0

	fmt.Printf("Fetching games for user '%s' from %s to %s\n", username, startDate.Format("Jan 2006"), endDate.Format("Jan 2006"))

	// --- Game Fetching Loop ---
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 1, 0) {
		year := d.Format("2006")
		month := d.Format("01")
		fmt.Printf("... checking %s/%s\n", month, year)
		gamesResponse, err := client.FetchPlayerGamesByMonth(username, year, month)
		if err != nil {
			log.Printf("Could not fetch games for %s/%s: %v", month, year, err)
			continue
		}
		if gamesResponse != nil && len(gamesResponse.Games) > 0 {
			allGames = append(allGames, gamesResponse.Games...)
			totalGamesFound += len(gamesResponse.Games)
		}
		time.Sleep(250 * time.Millisecond)
	}

	// --- Display Results ---
	fmt.Printf("\n--- Finished Fetching --- \n")
	fmt.Printf("Found a total of %d games for %s.\n\n", totalGamesFound, username)
	if totalGamesFound == 0 {
		return
	}
	listGames(allGames)

	// --- Interactive Game Selection ---
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nEnter a game number to select, or 'quit' to exit: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		gameNum, err := strconv.Atoi(input)
		if err != nil || gameNum < 1 || gameNum > len(allGames) {
			fmt.Println("Invalid number. Please enter a number from the list.")
			continue
		}

		// Enter the sub-menu for the selected game
		handleSelectedGame(reader, analyser, allGames[gameNum-1], gameNum)
		listGames(allGames) // Re-list games after returning from sub-menu
	}
}

// listGames prints the list of fetched games.
func listGames(games []api.Game) {
	fmt.Println("--- Games Found ---")
	for i, game := range games {
		endTime := time.Unix(game.EndTime, 0)
		fmt.Printf("[%d] %s vs %s (%s) - Played on %s\n",
			i+1, game.White.Username, game.Black.Username, game.TimeClass, endTime.Format("2006-01-02"))
	}
	fmt.Println("-------------------")
}

// handleSelectedGame provides options for a selected game (details, analyse).
func handleSelectedGame(reader *bufio.Reader, analyser *gameengine.StockfishAnalyser, game api.Game, gameNum int) {
	for {
		fmt.Printf("\nSelected Game %d: %s vs %s\n", gameNum, game.White.Username, game.Black.Username)
		fmt.Print("Enter command ('details', 'analyse', 'back'): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "details":
			displayGameDetails(game, gameNum)
		case "analyse":
			analyseGameMoves(analyser, game)
		case "back":
			return
		default:
			fmt.Println("Invalid command.")
		}
	}
}

// displayGameDetails shows detailed information for a selected game.
func displayGameDetails(game api.Game, index int) {
	endTime := time.Unix(game.EndTime, 0)
	fmt.Printf("\n--- Game Details (%d) ---\n", index)
	fmt.Printf("URL: %s\n", game.URL)
	fmt.Printf("Date: %s\n", endTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Result: White: %s, Black: %s\n", game.White.Result, game.Black.Result)
	fmt.Println("--- PGN ---")
	fmt.Println(game.PGN)
	fmt.Println("-------------")
}

// analyseGameMoves triggers the stockfish analysis and prints the results.
func analyseGameMoves(analyser *gameengine.StockfishAnalyser, game api.Game) {
	fmt.Println("\nAnalysing game... this may take a moment.")
	analysis, err := analyser.AnalyseGame(game)
	if err != nil {
		log.Printf("Error during analysis: %v", err)
		return
	}

	fmt.Println("\n--- Move Analysis ---")
	fmt.Println("Move | White              | Black              | Eval")
	fmt.Println("-----------------------------------------------------")
	for i := 0; i < len(analysis); i += 2 {
		whiteMove := analysis[i]
		var blackMoveStr string
		if i+1 < len(analysis) {
			blackMove := analysis[i+1]
			blackMoveStr = fmt.Sprintf("%-20s", blackMove.Move)
		} else {
			blackMoveStr = fmt.Sprintf("%-20s", "")
		}

		fmt.Printf("%-4d | %-20s | %s | %s\n",
			whiteMove.MoveNumber,
			whiteMove.Move,
			blackMoveStr,
			whiteMove.EvaluationText,
		)
	}
	fmt.Println("---------------------")
}

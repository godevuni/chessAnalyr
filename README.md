# Chess Analyser

A command-line tool to fetch and analyse Chess.com games using the Stockfish chess engine.

## Features

- Fetches games for a given Chess.com username over a specified date range.
- Lists games with player names, time class, and date.
- View detailed information and PGN for each game.
- Analyse each move using Stockfish and display evaluation in pawns.

## Requirements

- Go 1.24 or later
- [Stockfish](https://stockfishchess.org/download/) chess engine (download and note the path)
- Internet connection (to fetch games from Chess.com)

## Installation

1. Clone this repository:
    ```sh
    git clone https://github.com/yourusername/chessAnalyserFree.git
    cd chessAnalyserFree
    ```

2. Install Go dependencies:
    ```sh
    go mod tidy
    ```

3. Ensure Stockfish is installed and note its path (e.g., `/usr/local/bin/stockfish`).

## Usage

Run the program with:

```sh
go run . <username> <start_YYYY-MM> <end_YYYY-MM> <path_to_stockfish>
```

**Example:**
```sh
go run . hikaru 2022-10 2023-01 /usr/local/bin/stockfish
```

- `<username>`: Chess.com username
- `<start_YYYY-MM>`: Start date (e.g., 2022-10)
- `<end_YYYY-MM>`: End date (e.g., 2023-01)
- `<path_to_stockfish>`: Path to your Stockfish executable

## Interactive Commands

After fetching games, you can:

- Enter a game number to select a game.
- In the game menu:
    - `details`: Show game details and PGN.
    - `analyse`: Analyse the game move by move with Stockfish.
    - `back`: Return to the games list.
- `quit`: Exit the program.

## Project Structure

- `main.go`: Main CLI logic.
- `api/ChessComGame.go`: Chess.com API client and game data structures.
- `gameEngine/StockfishAnalyser.go`: Stockfish engine integration and move analysis.
- `gameFetch/`: (For future expansion, currently not used in main flow.)

## License

MIT License


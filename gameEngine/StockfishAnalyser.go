package gameengine

import (
	"bufio"
	"chessAnalyserFree/api"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/notnil/chess"
)

// MoveAnalysis holds the evaluation for a single move.
type MoveAnalysis struct {
	MoveNumber     int
	Move           string
	Evaluation     float64 // Evaluation in pawns (+ for white, - for black)
	EvaluationText string  // e.g., "+1.23" or "-0.54"
}

// StockfishAnalyser manages the communication with the Stockfish engine.
type StockfishAnalyser struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reader *bufio.Reader
}

// NewStockfishAnalyser starts the Stockfish process.
// You must provide the path to the Stockfish executable.
func NewStockfishAnalyser(stockfishPath string) (*StockfishAnalyser, error) {
	cmd := exec.Command(stockfishPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start stockfish: %w. Is the path correct?", err)
	}

	analyser := &StockfishAnalyser{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		reader: bufio.NewReader(stdout),
	}

	// Initialize UCI protocol
	if err := analyser.sendCommand("uci"); err != nil {
		return nil, err
	}
	// Wait for 'uciok'
	if _, err := analyser.readUntil("uciok"); err != nil {
		return nil, err
	}
	// Wait for 'readyok'
	if err := analyser.sendCommand("isready"); err != nil {
		return nil, err
	}
	if _, err := analyser.readUntil("readyok"); err != nil {
		return nil, err
	}

	return analyser, nil
}

// sendCommand sends a command string to the Stockfish process.
func (s *StockfishAnalyser) sendCommand(command string) error {
	_, err := fmt.Fprintln(s.stdin, command)
	return err
}

// readUntil reads from Stockfish's stdout until a line containing the specified text is found.
func (s *StockfishAnalyser) readUntil(contains string) (string, error) {
	var output string
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		output += line
		if strings.Contains(line, contains) {
			return output, nil
		}
	}
}

// AnalyseGame takes a game object and returns an analysis for each move.
func (s *StockfishAnalyser) AnalyseGame(game api.Game) ([]MoveAnalysis, error) {
	// --- CORRECTED PGN PARSING LOGIC ---
	// Use chess.PGN to create a parser, then pass it to chess.NewGame.
	pgnReader := strings.NewReader(game.PGN)
	pgnParser, err := chess.PGN(pgnReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create PGN parser: %w", err)
	}
	// Create a new game by applying the PGN data.
	parsedGame := chess.NewGame(pgnParser)
	// --- END OF CORRECTION ---

	// Create a separate game state from the initial position to replay moves for analysis.
	gameLogic := chess.NewGame()
	var analysis []MoveAnalysis

	// Regex to find the centipawn score from Stockfish's output.
	scoreRegex := regexp.MustCompile(`score cp (-?\d+)`)

	// Iterate through all moves that were actually played in the game.
	for i, move := range parsedGame.Moves() {
		// Get the board state (FEN) *before* the current move is made.
		fen := gameLogic.FEN()

		// Tell Stockfish to analyze this position.
		s.sendCommand(fmt.Sprintf("position fen %s", fen))
		// Analyze for 500 milliseconds. Increase for better accuracy.
		s.sendCommand("go movetime 500")

		// Find the line containing the evaluation score.
		output, err := s.readUntil("bestmove")
		if err != nil {
			return nil, fmt.Errorf("error reading from stockfish: %w", err)
		}

		var centipawns int
		matches := scoreRegex.FindStringSubmatch(output)
		if len(matches) > 1 {
			cp, _ := strconv.Atoi(matches[1])
			centipawns = cp
		}

		// Convert centipawns to pawn units.
		pawnEvaluation := float64(centipawns) / 100.0

		analysis = append(analysis, MoveAnalysis{
			MoveNumber:     (i / 2) + 1,
			Move:           move.String(),
			Evaluation:     pawnEvaluation,
			EvaluationText: fmt.Sprintf("%+.2f", pawnEvaluation),
		})

		// Apply the move to our logical board to advance to the next position.
		if err := gameLogic.Move(move); err != nil {
			return nil, fmt.Errorf("invalid move found in PGN: %w", err)
		}
	}

	return analysis, nil
}

// Close gracefully terminates the Stockfish process.
func (s *StockfishAnalyser) Close() {
	s.sendCommand("quit")
	s.cmd.Wait()
	s.stdin.Close()
	s.stdout.Close()
}

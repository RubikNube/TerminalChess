// Package engine provides an interface to the Stockfish chess engine.
package engine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type EngineConfig struct {
	Name        string                 `json:"name"`
	Threads     int                    `json:"threads"`
	Depth       int                    `json:"depth"`
	Automove    bool                   `json:"automove"`
	EngineColor string                 `json:"engineColor"`
	Path        string                 `json:"path"`
	Options     map[string]interface{} `json:"options"`
}

// Engine wraps a Stockfish process.
type Engine struct {
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Scanner
}

var (
	configFilePath = "engine.json" // Default path for engine configuration file
	engineConfig   EngineConfig
	loadedEngine   *Engine // Singleton instance of the engine
)

func Initialize(engineConfigPath string) error {
	// Load engine configuration
	cfg, err := loadEngineConfig(engineConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load engine config: %w", err)
	}
	engineConfig = cfg

	startStockfishWithOptions(cfg.Path, cfg.Options)

	return nil
}

func loadEngineConfig(path string) (EngineConfig, error) {
	var cfg EngineConfig
	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&cfg)
	return cfg, err
}

// StartStockfishWithOptions launches Stockfish and sets UCI options from a map.
func startStockfishWithOptions(path string, options map[string]interface{}) error {
	cmd := exec.Command(path)
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	if loadedEngine == nil {
		loadedEngine = &Engine{
			cmd:    cmd,
			stdin:  bufio.NewWriter(stdinPipe),
			stdout: bufio.NewScanner(stdoutPipe),
		}
	}
	// Initialize UCI
	SendCommand("uci")
	readUntil("uciok")

	// Set UCI options
	for k, v := range options {
		switch val := v.(type) {
		case bool:
			SendCommand(fmt.Sprintf("setoption name %s value %v", k, val))
		case float64:
			// JSON numbers are float64, but Stockfish expects int for most options
			SendCommand(fmt.Sprintf("setoption name %s value %d", k, int(val)))
		default:
			SendCommand(fmt.Sprintf("setoption name %s value %v", k, val))
		}
	}

	// Wait for Stockfish to process options
	SendCommand("isready")
	readUntil("readyok")

	return nil
}

// SendCommand sends a command to Stockfish.
func SendCommand(cmd string) {
	loadedEngine.stdin.WriteString(cmd + "\n")
	loadedEngine.stdin.Flush()
}

// readUntil reads lines until a line contains the given substring.
func readUntil(substr string) {
	for loadedEngine.stdout.Scan() {
		line := loadedEngine.stdout.Text()
		if strings.Contains(line, substr) {
			break
		}
	}
}

// GetBestMove returns the best move for a given FEN position.
func GetBestMove(fen string, depth int) (string, error) {
	SendCommand("position fen " + fen)
	SendCommand(fmt.Sprintf("go depth %d", depth))
	for loadedEngine.stdout.Scan() {
		line := loadedEngine.stdout.Text()
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
			break
		}
	}
	return "", fmt.Errorf("no bestmove found")
}

// Close terminates the Stockfish process.
func Close() {
	if loadedEngine.cmd != nil && loadedEngine.cmd.Process != nil {
		loadedEngine.cmd.Process.Kill()
	}
}

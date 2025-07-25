// Package engine provides an interface to the Stockfish chess engine.
package engine

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// Engine wraps a Stockfish process.
type Engine struct {
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Scanner
}

// StartStockfish launches the Stockfish process and returns an Engine.
func StartStockfish() (*Engine, error) {
	cmd := exec.Command("stockfish")
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	engine := &Engine{
		cmd:    cmd,
		stdin:  bufio.NewWriter(stdinPipe),
		stdout: bufio.NewScanner(stdoutPipe),
	}
	// Initialize UCI
	engine.SendCommand("uci")
	engine.ReadUntil("uciok")
	return engine, nil
}

// SendCommand sends a command to Stockfish.
func (e *Engine) SendCommand(cmd string) {
	e.stdin.WriteString(cmd + "\n")
	e.stdin.Flush()
}

// ReadUntil reads lines until a line contains the given substring.
func (e *Engine) ReadUntil(substr string) {
	for e.stdout.Scan() {
		line := e.stdout.Text()
		if strings.Contains(line, substr) {
			break
		}
	}
}

// GetBestMove returns the best move for a given FEN position.
func (e *Engine) GetBestMove(fen string, depth int) (string, error) {
	e.SendCommand("position fen " + fen)
	e.SendCommand(fmt.Sprintf("go depth %d", depth))
	for e.stdout.Scan() {
		line := e.stdout.Text()
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
func (e *Engine) Close() {
	if e.cmd != nil && e.cmd.Process != nil {
		e.cmd.Process.Kill()
	}
}

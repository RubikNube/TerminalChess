package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RubikNube/TerminalChess/pkg/gui"
)

func TestLoadConfig_Success(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "test_config.json")
	cfgData := `{"keybindings":{"moveLeft":"h","moveRight":"l","moveUp":"k","moveDown":"j","pick":"p","drop":"d","quit":"q","reset":"r","toggleHistory":"t","switchBoard":"s"}}`
	if err := os.WriteFile(tmpFile, []byte(cfgData), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg, err := loadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if cfg.Keybindings["moveLeft"] != "h" {
		t.Errorf("Expected moveLeft to be 'h', got %v", cfg.Keybindings["moveLeft"])
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := loadConfig("nonexistent.json")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "bad_config.json")
	if err := os.WriteFile(tmpFile, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}
	defer os.Remove(tmpFile)

	_, err := loadConfig(tmpFile)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestChessBoardInitialization(t *testing.T) {
	board := gui.NewChessBoard()
	if len(board) != 8 {
		t.Errorf("Expected board to have 8 rows, got %d", len(board))
	}
	for i, row := range board {
		if len(row) != 8 {
			t.Errorf("Row %d: expected 8 columns, got %d", i, len(row))
		}
	}
}

func TestMovePiece_InvalidMove(t *testing.T) {
	board := gui.NewChessBoard()
	// Try to move from an empty square
	ok := board.MovePiece(3, 3, 4, 4, gui.White)
	if ok {
		t.Error("Expected move to fail from empty square")
	}
}

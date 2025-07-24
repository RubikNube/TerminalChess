package history

import (
	"testing"

	"github.com/corentings/chess"
)

func TestAddMoveAndGetHistory(t *testing.T) {
	ClearHistory()
	AddMove("e2e4")
	AddMove("e7e5")
	h := GetHistory()
	if len(h) != 2 {
		t.Fatalf("Expected 2 moves, got %d", len(h))
	}
	if h[0] != "e2e4" || h[1] != "e7e5" {
		t.Errorf("Unexpected move history: %v", h)
	}
}

func TestClearHistory(t *testing.T) {
	AddMove("d2d4")
	ClearHistory()
	h := GetHistory()
	if len(h) != 0 {
		t.Errorf("Expected history to be empty after clear, got %v", h)
	}
}

func TestGetMoveHistorySAN_ValidMoves(t *testing.T) {
	ClearHistory()
	AddMove("e2e4")
	AddMove("e7e5")
	AddMove("g1f3")
	AddMove("b8c6")
	san := GetMoveHistorySAN()
	if len(san) != 2 {
		t.Errorf("Expected 2 SAN lines, got %d", len(san))
	}
	if san[0] == "" || san[1] == "" {
		t.Errorf("Expected non-empty SAN lines, got %v", san)
	}
}

func TestGetMoveHistorySAN_InvalidMove(t *testing.T) {
	ClearHistory()
	AddMove("invalid")
	san := GetMoveHistorySAN()
	if len(san) != 1 {
		t.Errorf("Expected 1 SAN line for invalid move, got %d", len(san))
	}
	if san[0] != "1. invalid" {
		t.Errorf("Expected SAN to show raw move for invalid input, got %v", san[0])
	}
}

func TestIsInCheck_NoCheck(t *testing.T) {
	game := chess.NewGame()
	if IsInCheck(game) {
		t.Error("Expected no check in starting position")
	}
}

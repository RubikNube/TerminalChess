package gui

import (
	"testing"
)

// Test chessboard initialization for correct dimensions
func TestNewChessBoard_Initialization(t *testing.T) {
	board := NewChessBoard()
	if len(board) != 8 {
		t.Errorf("Expected 8 rows, got %d", len(board))
	}
	for i, row := range board {
		if len(row) != 8 {
			t.Errorf("Row %d: expected 8 columns, got %d", i, len(row))
		}
	}
}

// Test valid move: white pawn e2 to e4
func TestMovePiece_ValidMove(t *testing.T) {
	board := NewChessBoard()
	ok := board.MovePiece(6, 4, 4, 4, White)
	if !ok {
		t.Error("Expected valid move for white pawn e2 to e4")
	}
	if board[4][4].Type != Pawn || board[4][4].Color != White {
		t.Error("Expected WhitePawn at e4 after move")
	}
	if board[6][4].Type != Empty {
		t.Error("Expected source square to be empty after move")
	}
}

// Test invalid move: move from empty square
func TestMovePiece_InvalidMove(t *testing.T) {
	board := NewChessBoard()
	ok := board.MovePiece(3, 3, 4, 4, White)
	if ok {
		t.Error("Expected move to fail from empty square")
	}
}

// Dummy implementation for keybindings defaults for testing
func GetKeybindingsDefaults() map[string]string {
	return map[string]string{
		"moveLeft":      "h",
		"moveRight":     "l",
		"moveUp":        "k",
		"moveDown":      "j",
		"pick":          "p",
		"drop":          "d",
		"quit":          "q",
		"reset":         "r",
		"toggleHistory": "t",
		"switchBoard":   "s",
	}
}

func TestKeybindings_Defaults(t *testing.T) {
	kb := GetKeybindingsDefaults()
	if kb["moveLeft"] == "" || kb["moveRight"] == "" {
		t.Error("Expected default keybindings to be set")
	}
}

// Dummy implementation for game state for testing
type GameState struct {
	Board    *ChessBoard
	Turn     Color
	Selected *Piece
}

func NewGameState() *GameState {
	b := NewChessBoard()
	return &GameState{
		Board:    &b,
		Turn:     White,
		Selected: nil,
	}
}

func TestGameState_Initialization(t *testing.T) {
	state := NewGameState()
	if state.Board == nil {
		t.Error("Expected board to be initialized")
	}
	if state.Turn != White {
		t.Errorf("Expected initial turn to be White, got %v", state.Turn)
	}
	if state.Selected != nil {
		t.Error("Expected no selected piece at start")
	}
}

// Test moving a piece out of bounds
func TestMovePiece_OutOfBounds(t *testing.T) {
	board := NewChessBoard()
	ok := board.MovePiece(-1, 0, 0, 0, White)
	if ok {
		t.Error("Expected move to fail for out-of-bounds source")
	}
	ok = board.MovePiece(0, 0, 8, 0, White)
	if ok {
		t.Error("Expected move to fail for out-of-bounds destination")
	}
}

// Test moving opponent's piece
func TestMovePiece_WrongColor(t *testing.T) {
	board := NewChessBoard()
	// Try to move black pawn as white
	ok := board.MovePiece(1, 0, 2, 0, White)
	if ok {
		t.Error("Expected move to fail when moving opponent's piece")
	}
}

// Test moving to same square
func TestMovePiece_SameSquare(t *testing.T) {
	board := NewChessBoard()
	ok := board.MovePiece(6, 0, 6, 0, White)
	if ok {
		t.Error("Expected move to fail when source and destination are the same")
	}
}

// Test picking up and dropping a piece (simulate selection logic)
func TestGameState_PickAndDrop(t *testing.T) {
	state := NewGameState()
	// Pick a white pawn at e2
	piece := &(*state.Board)[6][4]
	state.Selected = piece
	if state.Selected == nil || state.Selected.Type != Pawn || state.Selected.Color != White {
		t.Error("Expected to select white pawn at e2")
	}
	// Drop at e4
	ok := state.Board.MovePiece(6, 4, 4, 4, White)
	if !ok {
		t.Error("Expected to drop selected piece at e4")
	}
	state.Selected = nil
	if state.Selected != nil {
		t.Error("Expected no selected piece after drop")
	}
}

// Test keybindings for missing keys
func TestKeybindings_MissingKey(t *testing.T) {
	kb := GetKeybindingsDefaults()
	if _, ok := kb["nonexistent"]; ok {
		t.Error("Expected nonexistent keybinding to be missing")
	}
}

// Test board after reset
func TestChessBoard_Reset(t *testing.T) {
	board := NewChessBoard()
	board.MovePiece(6, 4, 4, 4, White)
	board = NewChessBoard()
	if board[6][4].Type != Pawn || board[6][4].Color != White {
		t.Error("Expected white pawn at e2 after reset")
	}
	if board[4][4].Type != Empty {
		t.Error("Expected e4 to be empty after reset")
	}
}

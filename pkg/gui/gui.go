// Package gui provides a simple text-based GUI for displaying a chess board.
package gui

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

// ChessBoard represents a simple 8x8 chess board.
type ChessBoard [8][8]rune

// NewChessBoard initializes a chess board with the standard starting position.
func NewChessBoard() ChessBoard {
	board := ChessBoard{}
	// White pieces
	board[0] = [8]rune{'R', 'N', 'B', 'Q', 'K', 'B', 'N', 'R'}
	board[1] = [8]rune{'P', 'P', 'P', 'P', 'P', 'P', 'P', 'P'}
	// Empty squares
	for i := 2; i < 6; i++ {
		for j := 0; j < 8; j++ {
			board[i][j] = '.'
		}
	}
	// Black pieces
	board[6] = [8]rune{'p', 'p', 'p', 'p', 'p', 'p', 'p', 'p'}
	board[7] = [8]rune{'r', 'n', 'b', 'q', 'k', 'b', 'n', 'r'}
	return board
}

// Cursor represents the position and selection state on the board.
type Cursor struct {
	Row      int
	Col      int
	Selected bool
}

// Move updates the cursor position by the given delta, clamped to board bounds.
func (c *Cursor) Move(dRow, dCol int) {
	nextRow := c.Row + dRow
	nextCol := c.Col + dCol
	if nextRow >= 0 && nextRow < 8 {
		c.Row = nextRow
	}
	if nextCol >= 0 && nextCol < 8 {
		c.Col = nextCol
	}
}

// RenderToView prints the chess board to a gocui.View, highlighting the cursor.
func (b ChessBoard) RenderToView(v *gocui.View, cursorRow, cursorCol int) {
	v.Clear()
	fmt.Fprintln(v, "  a b c d e f g h")
	for i := 0; i < 8; i++ {
		fmt.Fprintf(v, "%d ", 8-i)
		for j := 0; j < 8; j++ {
			cell := fmt.Sprintf("%c", b[i][j])
			if i == cursorRow && j == cursorCol {
				cell = "[" + cell + "]"
			}
			fmt.Fprintf(v, "%s ", cell)
		}
		fmt.Fprintf(v, "%d\n", 8-i)
	}
	fmt.Fprintln(v, "  a b c d e f g h")
}

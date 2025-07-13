// Package gui provides a simple text-based GUI for displaying a chess board.
package gui

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

// constants for the colored squares
const (
	WhiteSquare = "." // Light square
	BlackSquare = " " // Dark square
	EmptySquare = " " // Empty square
)

const (
	Black = iota
	White
)

const (
	King = iota
	Queen
	Rook
	Bishop
	Knight
	Pawn
)

// get rune for piece type
func pieceRune(pieceType int) rune {
	switch pieceType {
	case King:
		return '♚'
	case Queen:
		return '♛'
	case Rook:
		return '♜'
	case Bishop:
		return '♝'
	case Knight:
		return '♞'
	case Pawn:
		return '♟'
	default:
		return ' '
	}
}

type Piece struct {
	Color int // Black or White
	Type  int // King, Queen, Rook, Bishop, Knight, Pawn
}

// ChessBoard represents a simple 8x8 chess board.
type ChessBoard [8][8]Piece

// NewChessBoard initializes a chess board with the standard starting position.
func NewChessBoard() ChessBoard {
	board := ChessBoard{}
	// Initialize pawns
	for i := 0; i < 8; i++ {
		board[1][i] = Piece{Color: Black, Type: Pawn} // Black pawns
		board[6][i] = Piece{Color: White, Type: Pawn} // White pawns
	}
	// Initialize rooks
	board[0][0] = Piece{Color: Black, Type: Rook}
	board[0][7] = Piece{Color: Black, Type: Rook}
	board[7][0] = Piece{Color: White, Type: Rook}
	board[7][7] = Piece{Color: White, Type: Rook}
	// Initialize knights
	board[0][1] = Piece{Color: Black, Type: Knight}
	board[0][6] = Piece{Color: Black, Type: Knight}
	board[7][1] = Piece{Color: White, Type: Knight}
	board[7][6] = Piece{Color: White, Type: Knight}
	// Initialize bishops
	board[0][2] = Piece{Color: Black, Type: Bishop}
	board[0][5] = Piece{Color: Black, Type: Bishop}
	board[7][2] = Piece{Color: White, Type: Bishop}
	board[7][5] = Piece{Color: White, Type: Bishop}
	// Initialize queens
	board[0][3] = Piece{Color: Black, Type: Queen} // Black queen
	board[7][3] = Piece{Color: White, Type: Queen} // White queen
	// Initialize kings
	board[0][4] = Piece{Color: Black, Type: King} // Black king
	board[7][4] = Piece{Color: White, Type: King} // White king
	// Initialize empty squares
	for i := 2; i < 6; i++ {
		for j := 0; j < 8; j++ {
			board[i][j] = Piece{Color: -1, Type: -1} // Empty square
		}
	}
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

// RenderToView prints the chess board to a gocui.View, showing labels only on the top and left.
// Draws black and white squares using reverse video and bold formatting.
// Ensures quadratic rendering by padding each square to two characters wide.
func (b ChessBoard) RenderToView(v *gocui.View, cursorRow, cursorCol int) {
	v.Clear()
	// Top column labels, aligned with board
	fmt.Fprint(v, " ")
	for col := 0; col < 8; col++ {
		fmt.Fprintf(v, " %c", 'a'+col)
	}
	fmt.Fprintln(v)
	for i := 0; i < 8; i++ {
		// Row label on the left
		fmt.Fprintf(v, "%d ", 8-i)
		for j := 0; j < 8; j++ {
			piece := b[i][j]
			var fgColor, bgColor string

			// Determine piece color
			if piece.Color == Black {
				fgColor = "\033[31m" // Black text (bold)
			} else if piece.Color == White {
				fgColor = "\033[34m" // White text (bold)
			} else {
				fgColor = "\033[0m" // Default text color for empty squares
			}

			// Determine square color
			if (i+j)%2 == 0 {
				bgColor = "\033[47m" // White square background
			} else {
				bgColor = "\033[40m" // Dark gray for black square background
			}

			// Cursor highlight: underline
			cursorAttr := ""
			if i == cursorRow && j == cursorCol {
				cursorAttr = "\033[4m"
			}

			// Get the piece rune
			pieceRune := pieceRune(piece.Type)
			cell := fmt.Sprintf("%s%s%s%-2c\033[0m", fgColor, bgColor, cursorAttr, pieceRune)
			fmt.Fprint(v, cell)
		}
		fmt.Fprintln(v)
	}
}

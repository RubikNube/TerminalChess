// Package gui provides a simple text-based GUI for displaying a chess board.
package gui

import (
	"fmt"
	"strings"

	"github.com/corentings/chess"
	"github.com/jroimartin/gocui"
)

type SquareColor string

// constants for the colored squares
const (
	WhiteSquare SquareColor = "." // Light square
	BlackSquare             = " " // Dark square
	EmptySquare             = " " // Empty square
)

type Color int

const (
	Black Color = iota
	White
	Undefined = -1 // Represents an undefined color
)

type PieceType int

const (
	King PieceType = iota
	Queen
	Rook
	Bishop
	Knight
	Pawn
	Empty = -1 // Represents an empty square
)

type Piece struct {
	Color Color     // Black or White
	Type  PieceType // King, Queen, Rook, Bishop, Knight, Pawn
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
			board[i][j] = Piece{Color: Undefined, Type: Empty} // Empty square
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

// MovePiece moves a piece from (fromRow, fromCol) to (toRow, toCol) if the move is legal.
func (b *ChessBoard) MovePiece(fromRow, fromCol, toRow, toCol int, turn Color) bool {
	// Export current board to FEN, with correct turn
	fen := b.ToFEN(turn)
	chessFen, err := chess.FEN(fen)
	if err != nil {
		return false
	}
	game := chess.NewGame(chessFen)
	moveStr := fmt.Sprintf("%c%d%c%d", 'a'+fromCol, 8-fromRow, 'a'+toCol, 8-toRow)

	// Handle pawn promotion (promote to queen by default if moving to last rank)
	piece := b[fromRow][fromCol]
	if piece.Type == Pawn && (toRow == 0 || toRow == 7) {
		moveStr += "q"
	}

	move, err := chess.UCINotation{}.Decode(game.Position(), moveStr)
	if err != nil {
		return false
	}
	if err := game.Move(move); err != nil {
		return false
	}

	// Synchronize board with chess engine's position
	newBoard := game.Position().Board()
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			sq := chess.Square((7-i)*8 + j)
			p := newBoard.Piece(sq)
			if p == chess.NoPiece {
				b[i][j] = Piece{Color: Undefined, Type: Empty}
			} else {
				var color Color
				if p.Color() == chess.White {
					color = White
				} else {
					color = Black
				}
				var typ PieceType
				switch p.Type() {
				case chess.King:
					typ = King
				case chess.Queen:
					typ = Queen
				case chess.Rook:
					typ = Rook
				case chess.Bishop:
					typ = Bishop
				case chess.Knight:
					typ = Knight
				case chess.Pawn:
					typ = Pawn
				}
				b[i][j] = Piece{Color: color, Type: typ}
			}
		}
	}
	return true
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
func (b ChessBoard) RenderToView(v *gocui.View, cursorRow, cursorCol int, selected bool, selectedRow, selectedCol int) {
	v.Clear()
	artHeight := 7
	artWidth := 7
	// Top column labels, aligned with board
	squareWidth := artWidth*2 + 2 // doubled chars + 2 spaces padding
	fmt.Fprint(v, "  ")
	for col := 0; col < 8; col++ {
		label := fmt.Sprintf("%c", 'a'+col)
		pad := (squareWidth - len(label)) / 2
		fmt.Fprint(v, strings.Repeat(" ", pad))
		fmt.Fprint(v, label)
		fmt.Fprint(v, strings.Repeat(" ", squareWidth-pad-len(label)))
	}
	fmt.Fprintln(v)
	for i := 0; i < 8; i++ {
		for line := 0; line < artHeight; line++ {
			if line == 0 {
				fmt.Fprintf(v, "%d ", 8-i)
			} else {
				fmt.Fprint(v, "  ")
			}
			for j := 0; j < 8; j++ {
				piece := b[i][j]
				art := asciiPieces[piece.Type][piece.Color]
				cell := art[line]
				var fgColor, bgColor string
				reset := "\033[0m"

				// Determine piece color
				switch piece.Color {
				case Black:
					fgColor = "\033[31m"
				case White:
					fgColor = "\033[34m"
				default:
					fgColor = "\033[0m"
				}

				// Determine square color
				if (i+j)%2 == 0 {
					bgColor = "\033[47m"
				} else {
					bgColor = "\033[40m"
				}
				// Cursor highlight: underline + yellow background
				cursorAttr := ""
				if i == cursorRow && j == cursorCol {
					cursorAttr = "\033[4m\033[43m"
				}

				// Selected piece highlight: reverse video
				if selected && i == selectedRow && j == selectedCol {
					cursorAttr += "\033[7m"
				}
				// Double each character in the cell for horizontal scaling
				for _, ch := range cell {
					fmt.Fprintf(v, "%s%s%s%c%s", fgColor, bgColor, cursorAttr, ch, reset)
				}
			}
			fmt.Fprintln(v)
		}
	}
}

// ToFEN exports the ChessBoard to a FEN string (supports only piece placement, tracks turn, no castling/en passant).
func (b ChessBoard) ToFEN(turn Color) string {
	fen := ""
	for i := 0; i < 8; i++ {
		empty := 0
		for j := 0; j < 8; j++ {
			p := b[i][j]
			if p.Type == Empty {
				empty++
			} else {
				if empty > 0 {
					fen += fmt.Sprintf("%d", empty)
					empty = 0
				}
				var c rune
				switch p.Type {
				case King:
					c = 'k'
				case Queen:
					c = 'q'
				case Rook:
					c = 'r'
				case Bishop:
					c = 'b'
				case Knight:
					c = 'n'
				case Pawn:
					c = 'p'
				}
				if p.Color == White {
					c -= 32 // uppercase for white
				}
				fen += string(c)
			}
		}
		if empty > 0 {
			fen += fmt.Sprintf("%d", empty)
		}
		if i < 7 {
			fen += "/"
		}
	}
	// Use turn to set whose move it is
	turnStr := "w"
	if turn == Black {
		turnStr = "b"
	}
	// No castling, no en passant, fullmove 1, halfmove 0
	return fen + " " + turnStr + " - - 0 1"
}

//	                                          www www  _+_ _+_
//	 _   _    ,^.  ,^.   o    o    uuuu uuuu  \ / \#/  \ / \#/
//	( ) (#)  (  '\(##'\ ( /) (#/)  |  | |##|  ( ) (#)  ( ) (#)
//	/ \ /#\  |  \ |##\  /  \ /##\  /  \ /##\  / \ /#\  / \ /#\
//	=== ===  ==== ====  ==== ====  ==== ====  === ===  === ===PhS
//
// ASCII art for each piece type and color
var asciiPieces = map[PieceType]map[Color][]string{
	King: {
		White: {
			"              ",
			"      ██      ",
			"    ██████    ",
			"      ██      ",
			"     ████     ",
			"     ████     ",
			"   ████████   ",
		},
		Black: {
			"             ",
			"     ██      ",
			"   ██████    ",
			"     ██      ",
			"    ████     ",
			"    ████     ",
			"  ████████   ",
		},
	},
	Queen: {
		White: {
			"             ",
			"     ██      ",
			"    ████     ",
			"     ██      ",
			"    ████     ",
			"     ██      ",
			"   ██████    ",
		},
		Black: {
			"             ",
			"     ██      ",
			"    ████     ",
			"     ██      ",
			"    ████     ",
			"     ██      ",
			"   ██████    ",
		},
	},
	Rook: {
		White: {
			"             ",
			"             ",
			"    ████     ",
			"    █ ██     ",
			"    ████     ",
			"    ████     ",
			"   ██████    ",
		},
		Black: {
			"             ",
			"             ",
			"    ████     ",
			"    █ ██     ",
			"    ████     ",
			"    ████     ",
			"   ██████    ",
		},
	},
	Bishop: {
		White: {
			"             ",
			"             ",
			"     ██      ",
			"    ████     ",
			"    █ ██     ",
			"    ████     ",
			"   ██████    ",
		},
		Black: {
			"             ",
			"             ",
			"     ██      ",
			"    ████     ",
			"    █ ██     ",
			"    ████     ",
			"   ██████    ",
		},
	},
	Knight: {
		White: {
			"             ",
			"             ",
			"     ██      ",
			"    ████     ",
			"     ███     ",
			"    ███      ",
			"   ██████    ",
		},
		Black: {
			"             ",
			"             ",
			"     ██      ",
			"    ████     ",
			"     ███     ",
			"    ███      ",
			"   ██████    ",
		},
	},
	Pawn: {
		White: {
			"             ",
			"             ",
			"             ",
			"             ",
			"     ██      ",
			"    ████     ",
			"   ██████    ",
		},
		Black: {
			"             ",
			"             ",
			"             ",
			"             ",
			"     ██      ",
			"    ████     ",
			"   ██████    ",
		},
	},
	Empty: {
		White: {
			"             ",
			"             ",
			"             ",
			"             ",
			"             ",
			"             ",
			"             ",
			"             ",
		},
		Black: {
			"             ",
			"             ",
			"             ",
			"             ",
			"             ",
			"             ",
			"             ",
			"             ",
		},
		Undefined: {
			"             ",
			"             ",
			"             ",
			"             ",
			"             ",
			"             ",
			"             ",
			"             ",
		},
	},
}

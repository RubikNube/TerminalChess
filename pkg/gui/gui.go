// Package gui provides a simple text-based GUI for displaying a chess board.
package gui

import (
	"fmt"
	"os"
	"strings"

	"github.com/RubikNube/TerminalChess/pkg/history"
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

// Cursor represents the position and selection state on the board.
type Cursor struct {
	Row      int
	Col      int
	Selected bool
}

// ChessBoard represents a simple 8x8 chess board.
type ChessBoard [8][8]Piece

var asciiPieces = map[PieceType]map[Color][]string{}
var BoardFlipped bool = false

// Set the en passant square if the last move was a double pawn move
var enPassantRow int = -1
var enPassantCol int = -1

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

// NewChessBoardFromFEN creates a ChessBoard from a FEN string (piece placement only).
func NewChessBoardFromFEN(fen string) ChessBoard {
	board := ChessBoard{}
	rows := strings.Split(fen, "/")
	for i := 0; i < 8 && i < len(rows); i++ {
		row := rows[i]
		col := 0
		for _, c := range row {
			if col >= 8 {
				break
			}
			switch {
			case c >= '1' && c <= '8':
				for k := 0; k < int(c-'0'); k++ {
					board[i][col] = Piece{Color: Undefined, Type: Empty}
					col++
				}
			default:
				var color Color
				if c >= 'A' && c <= 'Z' {
					color = White
				} else {
					color = Black
				}
				var typ PieceType
				switch strings.ToLower(string(c)) {
				case "k":
					typ = King
				case "q":
					typ = Queen
				case "r":
					typ = Rook
				case "b":
					typ = Bishop
				case "n":
					typ = Knight
				case "p":
					typ = Pawn
				default:
					typ = Empty
				}
				board[i][col] = Piece{Color: color, Type: typ}
				col++
			}
		}
	}
	return board
}

// MovePiece moves a piece from (fromRow, fromCol) to (toRow, toCol) if the move is legal.
// Now supports castling and en passant by allowing king, rook, and pawn moves as per chess rules.
func (b *ChessBoard) MovePiece(fromRow, fromCol, toRow, toCol int, turn Color) bool {
	// Bounds check
	if fromRow < 0 || fromRow > 7 || fromCol < 0 || fromCol > 7 ||
		toRow < 0 || toRow > 7 || toCol < 0 || toCol > 7 {
		return false
	}
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

	// Try normal move
	move, err := chess.UCINotation{}.Decode(game.Position(), moveStr)
	if err != nil {
		// Try castling if king moves two squares horizontally
		if piece.Type == King && fromRow == toRow && abs(fromCol-toCol) == 2 {
			var castleMove *chess.Move
			if toCol == 6 { // kingside
				castleMove, _ = chess.UCINotation{}.Decode(game.Position(), "e1g1")
				if turn == Black {
					castleMove, _ = chess.UCINotation{}.Decode(game.Position(), "e8g8")
				}
			} else if toCol == 2 { // queenside
				castleMove, _ = chess.UCINotation{}.Decode(game.Position(), "e1c1")
				if turn == Black {
					castleMove, _ = chess.UCINotation{}.Decode(game.Position(), "e8c8")
				}
			}
			if castleMove != nil && game.Move(castleMove) == nil {
				updateBoardFromGame(b, game)
				history.AddMove(castleMove.String())
				setEnPassantSquare(piece, fromRow, fromCol, toRow, toCol)
				return true
			}
		}
		// Try en passant if pawn moves diagonally to an empty square and en passant is available
		if piece.Type == Pawn && fromRow != toRow && fromCol != toCol && b[toRow][toCol].Type == Empty {
			epRow, epCol := GetEnPassantSquare()
			if toRow == epRow && toCol == epCol {
				// Perform en passant capture
				updateBoardFromGame(b, game)
				// Remove the captured pawn
				if piece.Color == White {
					b[toRow+1][toCol] = Piece{Type: Empty, Color: Undefined}
				} else {
					b[toRow-1][toCol] = Piece{Type: Empty, Color: Undefined}
				}
				moveStr := move.String()
				moveStr += " e.p."
				history.AddMove(moveStr)
				setEnPassantSquare(piece, fromRow, fromCol, toRow, toCol)
				return true
			}
			// fallback: try normal pawn capture (should fail if not en passant)
			move, err = chess.UCINotation{}.Decode(game.Position(), moveStr)
			if err == nil && game.Move(move) == nil {
				updateBoardFromGame(b, game)
				history.AddMove(move.String())
				setEnPassantSquare(piece, fromRow, fromCol, toRow, toCol)
				return true
			}
		}
		return false
	}
	if err := game.Move(move); err != nil {
		return false
	}

	updateBoardFromGame(b, game)
	history.AddMove(move.String())
	setEnPassantSquare(piece, fromRow, fromCol, toRow, toCol)
	return true
}

func setEnPassantSquare(piece Piece, fromRow, fromCol, toRow, toCol int) {
	// Only pawns moving two squares forward
	if piece.Type == Pawn && abs(fromRow-toRow) == 2 && fromCol == toCol {
		// Set en passant square to the square behind the moved pawn
		if piece.Color == White {
			enPassantRow = toRow + 1
			enPassantCol = toCol
		} else if piece.Color == Black {
			enPassantRow = toRow - 1
			enPassantCol = toCol
		}
	} else {
		enPassantRow = -1
		enPassantCol = -1
	}
}

func GetEnPassantSquare() (int, int) {
	// Return the current en passant square
	return enPassantRow, enPassantCol
}

// updateBoardFromGame updates the ChessBoard from the chess.Game position.
func updateBoardFromGame(b *ChessBoard, game *chess.Game) {
	newBoard := game.Position().Board()
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			sq := chess.Square((7-i)*8 + j)
			p := newBoard.Piece(sq)
			if p == chess.NoPiece {
				(*b)[i][j] = Piece{Color: Undefined, Type: Empty}
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
				(*b)[i][j] = Piece{Color: color, Type: typ}
			}
		}
	}
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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

func (b ChessBoard) RenderToView(v *gocui.View, cursorRow, cursorCol int, selected bool, selectedRow, selectedCol int) {
	b.RenderToViewFlipped(v, cursorRow, cursorCol, selected, selectedRow, selectedCol, BoardFlipped)
}

func (b ChessBoard) RenderToViewFlipped(v *gocui.View, cursorRow, cursorCol int, selected bool, selectedRow, selectedCol int, flipped bool) {
	v.Clear()
	artHeight := 7
	artWidth := 7
	// Top column labels, aligned with board
	squareWidth := artWidth*2 + 2 // doubled chars + 2 spaces padding
	fmt.Fprint(v, "  ")
	for col := 0; col < 8; col++ {
		var labelCol int
		if flipped {
			labelCol = 7 - col
		} else {
			labelCol = col
		}
		label := fmt.Sprintf("%c", 'a'+labelCol)
		pad := (squareWidth - len(label)) / 2
		fmt.Fprint(v, strings.Repeat(" ", pad))
		fmt.Fprint(v, label)
		fmt.Fprint(v, strings.Repeat(" ", squareWidth-pad-len(label)-2))
	}
	fmt.Fprintln(v)
	for i := 0; i < 8; i++ {
		var row int
		if flipped {
			row = 7 - i
		} else {
			row = i
		}
		for line := 0; line < artHeight; line++ {
			var rowLabel string
			if line == artHeight/2 {
				rowLabel = fmt.Sprintf("%d ", 8-row)
			} else {
				rowLabel = "  "
			}
			fmt.Fprint(v, rowLabel)
			for j := 0; j < 8; j++ {
				var col int
				if flipped {
					col = 7 - j
				} else {
					col = j
				}
				piece := b[row][col]
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
				if (row+col)%2 == 0 {
					bgColor = "\033[47m"
				} else {
					bgColor = "\033[40m"
				}
				// Cursor highlight: yellow background only (no underline)
				cursorAttr := ""
				if row == cursorRow && col == cursorCol {
					cursorAttr = "\033[43m"
				}

				// Selected piece highlight: reverse video
				if selected && row == selectedRow && col == selectedCol {
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

// ToFEN exports the ChessBoard to a FEN string (supports only piece placement, tracks turn, and basic castling rights).
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
	// Compute castling rights (simple check: if king/rook are on original squares)
	castle := ""
	if b[7][4].Type == King && b[7][4].Color == White {
		if b[7][7].Type == Rook && b[7][7].Color == White {
			castle += "K"
		}
		if b[7][0].Type == Rook && b[7][0].Color == White {
			castle += "Q"
		}
	}
	if b[0][4].Type == King && b[0][4].Color == Black {
		if b[0][7].Type == Rook && b[0][7].Color == Black {
			castle += "k"
		}
		if b[0][0].Type == Rook && b[0][0].Color == Black {
			castle += "q"
		}
	}
	if castle == "" {
		castle = "-"
	}
	// Set en passant square if available
	ep := "-"
	if enPassantRow >= 0 && enPassantCol >= 0 && enPassantRow < 8 && enPassantCol < 8 {
		ep = fmt.Sprintf("%c%d", 'a'+enPassantCol, 8-enPassantRow)
	}
	// fullmove 1, halfmove 0
	return fen + " " + turnStr + " " + castle + " " + ep + " 0 1"
}

func LoadAsciiPieces(pieceFolder string) error {
	// Ensure the piece folder exists
	if _, err := os.Stat(pieceFolder); os.IsNotExist(err) {
		return fmt.Errorf("piece folder does not exist: %s", pieceFolder)
	}

	pieceFiles := map[PieceType]string{
		King:   pieceFolder + "/king.txt",
		Queen:  pieceFolder + "/queen.txt",
		Rook:   pieceFolder + "/rook.txt",
		Bishop: pieceFolder + "/bishop.txt",
		Knight: pieceFolder + "/knight.txt",
		Pawn:   pieceFolder + "/pawn.txt",
	}

	// Initialize asciiPieces map
	asciiPieces = make(map[PieceType]map[Color][]string)

	for typ, path := range pieceFiles {
		lines, err := readAsciiArtFile(path)
		if err != nil {
			return err
		}
		if asciiPieces[typ] == nil {
			asciiPieces[typ] = map[Color][]string{}
		}
		asciiPieces[typ][White] = lines
		asciiPieces[typ][Black] = lines // Optionally, use different files for Black
	}
	// Empty and Undefined pieces
	empty := make([]string, 7)
	for i := range empty {
		empty[i] = "              "
	}
	asciiPieces[Empty] = map[Color][]string{
		White:     empty,
		Black:     empty,
		Undefined: empty,
	}
	return nil
}

// GetMoveHistory returns the move history as a slice of strings.
func GetMoveHistory() []string {
	return history.GetHistory()
}

func readAsciiArtFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	// Remove empty trailing lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, nil
}

// ToggleBoardOrientation toggles the board orientation (flipped/unflipped).
func ToggleBoardOrientation() {
	BoardFlipped = !BoardFlipped
}

func PrintBoard(b ChessBoard) {
	// Unicode chess symbols (same for both colors)
	pieceSymbols := map[PieceType]string{
		King:   "♚",
		Queen:  "♛",
		Rook:   "♜",
		Bishop: "♝",
		Knight: "♞",
		Pawn:   "♟",
		Empty:  ".",
	}

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := b[i][j]
			// Determine ANSI background color
			bgColor := "\033[47m" // White
			if (i+j)%2 == 1 {
				bgColor = "\033[100m" // Gray for black square
			}
			// Determine piece color
			fgColor := ""
			switch piece.Color {
			case White:
				fgColor = "\033[93m"
			case Black:
				fgColor = "\033[94m"
			default:
				fgColor = "\033[90m" // Gray for undefined/empty
			}
			reset := "\033[0m"
			pieceChar := pieceSymbols[piece.Type]
			fmt.Printf("%s%s %s%s", bgColor, fgColor, pieceChar, reset)
		}
		fmt.Println()
	}
	println("")
}

// Package history provides functionality to track and manage the move history
package history

import (
	"fmt"
	"sync"

	"github.com/corentings/chess"
)

var (
	mu    sync.Mutex
	moves []string
)

// AddMove appends a move in algebraic notation to the history.
func AddMove(move string) {
	mu.Lock()
	defer mu.Unlock()
	moves = append(moves, move)
}

// GetHistory returns a copy of the move history.
func GetHistory() []string {
	mu.Lock()
	defer mu.Unlock()
	history := make([]string, len(moves))
	copy(history, moves)
	return history
}

// ClearHistory clears the move history.
func ClearHistory() {
	mu.Lock()
	defer mu.Unlock()
	moves = nil
}

// IsInCheck returns true if the side to move is in check in the given game position.
// This uses only the public API and custom logic for attack detection.
func IsInCheck(game *chess.Game) bool {
	pos := game.Position()
	board := pos.Board()
	turn := pos.Turn()
	opponent := turn.Other()

	// 1. Find the king's square for the side to move
	var kingSq chess.Square = chess.NoSquare
	for sq := chess.A1; sq <= chess.H8; sq++ {
		piece := board.Piece(sq)
		if piece != chess.NoPiece && piece.Type() == chess.King && piece.Color() == turn {
			kingSq = sq
			break
		}
	}
	if kingSq == chess.NoSquare {
		return false // king not found
	}

	// 2. Check for attacks from all opponent pieces
	for sq := chess.A1; sq <= chess.H8; sq++ {
		piece := board.Piece(sq)
		if piece == chess.NoPiece || piece.Color() != opponent {
			continue
		}
		switch piece.Type() {
		case chess.Pawn:
			// Pawns attack diagonally forward
			dir := 1
			if opponent == chess.White {
				dir = -1
			}
			// Check both diagonal squares
			for _, fileOffset := range []int{-1, 1} {
				attackedSq := chess.Square(int(sq) + dir*8 + fileOffset)
				if attackedSq >= chess.A1 && attackedSq <= chess.H8 && attackedSq == kingSq {
					// Make sure pawn is not wrapping around the board
					if abs(int(sq)%8-int(kingSq)%8) == 1 {
						return true
					}
				}
			}
		case chess.Knight:
			knightMoves := []int{15, 17, 6, 10, -15, -17, -6, -10}
			for _, offset := range knightMoves {
				attackedSq := chess.Square(int(sq) + offset)
				if attackedSq >= chess.A1 && attackedSq <= chess.H8 && attackedSq == kingSq {
					// Make sure move is valid (doesn't wrap around board)
					if isValidKnightMove(sq, attackedSq) {
						return true
					}
				}
			}
		case chess.Bishop, chess.Rook, chess.Queen:
			if canPieceReach(board, sq, kingSq, piece.Type()) {
				return true
			}
		case chess.King:
			if isKingAdjacent(sq, kingSq) {
				return true
			}
		}
	}
	return false
}

// Helper: absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Helper: check if two squares are adjacent (for king attacks)
func isKingAdjacent(from, to chess.Square) bool {
	df := abs(int(from)%8 - int(to)%8)
	dr := abs(int(from)/8 - int(to)/8)
	return (df <= 1 && dr <= 1) && from != to
}

// Helper: check if a knight move is valid (doesn't wrap around board)
func isValidKnightMove(from, to chess.Square) bool {
	df := abs(int(from)%8 - int(to)%8)
	dr := abs(int(from)/8 - int(to)/8)
	return (df == 1 && dr == 2) || (df == 2 && dr == 1)
}

// Helper: check if a sliding piece can reach the target square
func canPieceReach(board *chess.Board, from, to chess.Square, pt chess.PieceType) bool {
	df := int(to)%8 - int(from)%8
	dr := int(to)/8 - int(from)/8
	var stepF, stepR int
	switch pt {
	case chess.Bishop:
		if abs(df) != abs(dr) || df == 0 {
			return false
		}
		stepF = sign(df)
		stepR = sign(dr)
	case chess.Rook:
		if df != 0 && dr != 0 {
			return false
		}
		stepF = sign(df)
		stepR = sign(dr)
	case chess.Queen:
		if abs(df) == abs(dr) && df != 0 {
			stepF = sign(df)
			stepR = sign(dr)
		} else if (df == 0 && dr != 0) || (df != 0 && dr == 0) {
			stepF = sign(df)
			stepR = sign(dr)
		} else {
			return false
		}
	default:
		return false
	}
	// Step through the path
	f, r := int(from)%8, int(from)/8
	for {
		f += stepF
		r += stepR
		if f < 0 || f > 7 || r < 0 || r > 7 {
			return false
		}
		sq := chess.Square(r*8 + f)
		if sq == to {
			return true
		}
		if board.Piece(sq) != chess.NoPiece {
			return false
		}
	}
}

// Helper: sign function
func sign(x int) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

// GetMoveHistorySAN returns the move history as a slice of formatted strings in standard algebraic notation,
// including + for check and # for checkmate.
func GetMoveHistorySAN() []string {
	rawMoves := GetHistory()
	game := chess.NewGame()
	var lines []string
	for i := 0; i < len(rawMoves); i += 2 {
		moveNum := i/2 + 1
		var whiteSAN, blackSAN string

		// White move
		if i < len(rawMoves) {
			move, err := chess.UCINotation{}.Decode(game.Position(), rawMoves[i])
			if err == nil {
				whiteSAN = chess.AlgebraicNotation{}.Encode(game.Position(), move)
				game.Move(move)
				// Annotate with # for checkmate, + for check
				if game.Outcome() != chess.NoOutcome && game.Method() == chess.Checkmate {
					whiteSAN += "#"
				} else if IsInCheck(game) {
					whiteSAN += "+"
				}
			} else {
				whiteSAN = rawMoves[i]
			}
		}

		// Black move
		if i+1 < len(rawMoves) {
			move, err := chess.UCINotation{}.Decode(game.Position(), rawMoves[i+1])
			if err == nil {
				blackSAN = chess.AlgebraicNotation{}.Encode(game.Position(), move)
				game.Move(move)
				// Annotate with # for checkmate, + for check
				if game.Outcome() != chess.NoOutcome && game.Method() == chess.Checkmate {
					blackSAN += "#"
				} else if IsInCheck(game) {
					blackSAN += "+"
				}
			} else {
				blackSAN = rawMoves[i+1]
			}
		}

		if blackSAN != "" {
			lines = append(lines, fmt.Sprintf("%d. %s %s", moveNum, whiteSAN, blackSAN))
		} else {
			lines = append(lines, fmt.Sprintf("%d. %s", moveNum, whiteSAN))
		}
	}
	return lines
}

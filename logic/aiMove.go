package logic

import (
	"math"
	"math/rand"
)

// Convert slice to lookup map for O(1) membership checks
func makeBoardSet(boards []int32) map[int32]bool {
	set := make(map[int32]bool, len(boards))
	for _, idx := range boards {
		set[idx] = true
	}
	return set
}

// Heuristic: favor center cells
func getCellValue(globalIndex, boardSize int32) int32 {
	cellIndex := globalIndex % (boardSize * boardSize)
	r := float64(cellIndex / boardSize)
	c := float64(cellIndex % boardSize)
	center := float64(boardSize-1) / 2
	return int32(-math.Abs(r-center) - math.Abs(c-center))
}

// Produce all valid moves (global indexes)
func getValidMoves(boards []int32, boardSize, numberOfBoards int32) []int32 {
	set := makeBoardSet(boards)
	moves := []int32{}

	for b := int32(0); b < numberOfBoards; b++ {
		if IsBoardDead(b, boards, boardSize) {
			continue
		}
		boardOffset := b * boardSize * boardSize
		for i := int32(0); i < boardSize*boardSize; i++ {
			global := boardOffset + i
			if !set[global] {
				moves = append(moves, global)
			}
		}
	}

	// Sort by cell value (center bias)
	for i := 0; i < len(moves)-1; i++ {
		for j := i + 1; j < len(moves); j++ {
			if getCellValue(moves[j], boardSize) > getCellValue(moves[i], boardSize) {
				moves[i], moves[j] = moves[j], moves[i]
			}
		}
	}

	return moves
}

// Apply a move (returns new boards slice)
func updateBoards(boards []int32, move int32) []int32 {
	newBoards := make([]int32, len(boards)+1)
	copy(newBoards, boards)
	newBoards[len(boards)] = move
	return newBoards
}

// Main AI function: returns global move index (e.g. 13)
func GetAIMove(boards []int32, boardSize int32, numberOfBoards int32, difficulty int32) int32 {
	moves := getValidMoves(boards, boardSize, numberOfBoards)
	if len(moves) == 0 {
		return -1
	}

	optimalChance := float64(difficulty-1) / 4.0 // 0 @ diff=1 → 1 @ diff=5
	if rand.Float64() > optimalChance {
		return moves[rand.Intn(len(moves))]
	}

	// Count live boards
	liveCount := int32(0)
	for b := int32(0); b < numberOfBoards; b++ {
		if !IsBoardDead(b, boards, boardSize) {
			liveCount++
		}
	}

	// Split moves
	killing := []int32{}
	nonKilling := []int32{}
	for _, m := range moves {
		next := updateBoards(boards, m)
		boardIndex := m / (boardSize * boardSize)
		if IsBoardDead(boardIndex, next, boardSize) {
			killing = append(killing, m)
		} else {
			nonKilling = append(nonKilling, m)
		}
	}

	if liveCount%2 == 1 {
		// Winning position → leave boards alive
		if len(nonKilling) > 0 {
			return nonKilling[0]
		}
		return moves[0]
	} else {
		// Losing position → kill one to flip parity
		if len(killing) > 0 {
			return killing[rand.Intn(len(killing))]
		}
		return moves[0]
	}
}

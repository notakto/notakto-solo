package functions

func IsBoardDead(boardIndex int32, boards []int32, boardSize int32) bool {
	start := boardIndex * boardSize * boardSize
	end := start + boardSize*boardSize

	rowCount := make([]int32, boardSize)
	colCount := make([]int32, boardSize)
	var mainDiagCount, antiDiagCount int32

	for _, idx := range boards {
		if idx < start || idx >= end {
			continue
		}
		local := idx - start
		r := local / boardSize
		c := local % boardSize
		rowCount[r]++
		colCount[c]++
		if r == c {
			mainDiagCount++
		}
		if r+c == boardSize-1 {
			antiDiagCount++
		}
	}

	for i := range boardSize {
		if rowCount[i] == boardSize || colCount[i] == boardSize {
			return true
		}
	}

	if mainDiagCount == boardSize || antiDiagCount == boardSize {
		return true
	}

	return false
}

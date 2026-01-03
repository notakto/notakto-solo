package logic

import "math/rand/v2"

func CalculateRewards(NumberOfBoards int32, BoardSize int32, DifficultyLevel int32, win bool) (coinsReward int32, xpReward int32) {
	start := int32(0)
	end := int32(5)
	baseMultiplier := DifficultyLevel * NumberOfBoards * BoardSize
	xpMultiplier := rand.Int32N(end-start+1) + int32(6)
	coinMultiplier := rand.Int32N(end-start+1) + int32(1)
	if win {
		coinsReward = baseMultiplier * coinMultiplier
		xpReward = baseMultiplier * xpMultiplier
	} else {
		coinsReward = 0
		xpReward = baseMultiplier
	}
	return coinsReward, xpReward
}

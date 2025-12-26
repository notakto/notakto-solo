package types

// CreateGameRequest holds creation params
type CreateGameRequest struct {
	NumberOfBoards int32 `json:"numberOfBoards"`
	BoardSize      int32 `json:"boardSize"`
	Difficulty     int32 `json:"difficulty"`
}

type SignInResponse struct {
	Uid        string `json:"uid"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	ProfilePic string `json:"profile_pic"`
	NewAccount bool   `json:"new_account"`
}
type FirebaseTokenInfo struct {
	LocalID string `json:"localId"`
	Email   string `json:"email,omitempty"`
	Name    string `json:"displayName,omitempty"`
	Photo   string `json:"photoUrl,omitempty"`
}
type UpdatePlayerNameRequest struct {
	Name string `json:"name"`
}
type MakeMoveRequest struct {
	SessionID  string `json:"sessionId"`
	BoardIndex int32  `json:"boardIndex"`
	CellIndex  int32  `json:"cellIndex"`
}
type CreateGameResponse struct {
	SessionId      string  `json:"sessionId"`
	Uid            string  `json:"uid"`
	Boards         []int32 `json:"boards"`
	Winner         bool    `json:"winner"`
	BoardSize      int32   `json:"boardSize"`
	NumberOfBoards int32   `json:"numberOfBoards"`
	Difficulty     int32   `json:"difficulty"`
	Gameover       bool    `json:"gameover"`
	CreatedAt      string  `json:"createdAt"`
}
type MakeMoveResponse struct {
	Boards        []int32 `json:"boards"`
	Gameover      bool    `json:"gameover"`
	Winner        bool    `json:"winner"`
	CoinsRewarded int32   `json:"coinsRewarded"`
	XpRewarded    int32   `json:"xpRewarded"`
}
type QuitGameRequest struct {
	SessionID string `json:"sessionId"`
}
type QuitGameResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
type GetWalletResponse struct {
	Coins   int32  `json:"coins"`
	XP      int32  `json:"xp"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
type SkipMoveRequest struct {
	SessionID string `json:"sessionId"`
}
type SkipMoveResponse struct {
	Boards        []int32 `json:"boards"`
	Gameover      bool    `json:"gameover"`
	Winner        bool    `json:"winner"`
	CoinsRewarded int32   `json:"coinsRewarded"`
	XpRewarded    int32   `json:"xpRewarded"`
}
type UndoMoveRequest struct {
	SessionID string `json:"sessionId"`
}
type UndoMoveResponse struct {
	Boards []int32 `json:"boards"`
}
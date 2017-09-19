package models

type TournamentPlayer struct {
    ID              int `gorm:"primary_key;AUTO_INCREMENT"`
    Tournament      Tournament
    TournamentID    string
    Player          Player
    PlayerID        string
    Value           float64 `gorm:"size:11"`
    Role            string
    BackerForID     *string
}

package models

type Player struct {
    ID          string `gorm:"primary_key"`
    Points      float64 `gorm:"size:11"`
}

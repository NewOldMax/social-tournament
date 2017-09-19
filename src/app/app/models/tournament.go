package models

type Tournament struct {
    ID          string `gorm:"primary_key"`
    Deposit     float64 `gorm:"size:11"`
    Status      string
    Prize       float64 `gorm:"size:11"`
    Winners     *string `sql:"type:jsonb"`
}

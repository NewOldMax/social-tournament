package controllers

import (
    "github.com/revel/revel"
    gorm "github.com/Gr1N/revel-gorm/app"
    gormController "github.com/Gr1N/revel-gorm/app/controllers"
    "app/app/models"
)

func InitializeDB() {
    dbm := gorm.InitDB()
    dbm.AutoMigrate(&models.Player{}, &models.Tournament{}, &models.TournamentPlayer{})
}

func init() {
    revel.OnAppStart(func() {

        InitializeDB()
        revel.InterceptMethod((*gormController.TransactionalController).Begin, revel.BEFORE)
        revel.InterceptMethod((*gormController.TransactionalController).Commit, revel.AFTER)
        revel.InterceptMethod((*gormController.TransactionalController).Rollback, revel.FINALLY)
    })
}
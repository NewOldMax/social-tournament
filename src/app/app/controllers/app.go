package controllers

import (
	"github.com/revel/revel"
    "app/app/models"
)

type App struct {
    BasicController
}

func (c App) Index() revel.Result {
    return c.RenderMessage("it's alive")
}

func (c App) Reset() revel.Result {
    c.Txn.DropTableIfExists(&models.Player{}, &models.Tournament{}, &models.TournamentPlayer{})
    c.Txn.CreateTable(&models.Player{}, &models.Tournament{}, &models.TournamentPlayer{})
    return c.RenderMessage("database was resetted")
}

package controllers

import (
    "github.com/revel/revel"
    gorm "github.com/Gr1N/revel-gorm/app/controllers"
    "app/app/models"
)

type BasicController struct {
    gorm.TransactionalController
}

func (c BasicController) RenderMessage(message string) revel.Result {
    data := make(map[string]interface{})
    data["message"] = message
    return c.RenderJSON(data)
}

func (c BasicController) RenderErrorJSON(message string, code int) revel.Result {
    data := c.PrepareErrorJSON("Error", message, code)
    return c.RenderJSON(data)
}

func (c BasicController) PrepareErrorJSON(title string, message string, code int) interface{} {
    c.Response.Status = code
    data := make(map[string]interface{})
    data["Title"] = title
    data["Description"] = message
    return data
}

func (c BasicController) GetOrNotFoundPlayer() interface{} {
    playerId := c.Params.Query.Get("playerId")
    var player models.Player
    c.Txn.Where(&models.Player{ID: playerId}).First(&player)
    if player.ID == "" {
        data := c.PrepareErrorJSON("Not found", "Player with id '"+playerId+"' not found", 404)
        return data
    }
    return player
}

func (c BasicController) GetOrCreatePlayer() interface{} {
    playerId := c.Params.Query.Get("playerId")
    var player models.Player
    c.Txn.Where(&models.Player{ID: playerId}).First(&player)
    if player.ID == "" {
        c.Response.Status = 201
        player = models.Player{ID: playerId, Points: 0}
        c.Txn.Create(&player)
    }
    return player
}

func (c BasicController) ManagePlayerPoints(player models.Player, points float64, operation string) interface{} {
    if points <= 0 {
        return c.PrepareErrorJSON("Error", "Points amount should be positive", 400)
    }
    switch operation {
        case "take":
            if player.Points < points {
                return c.PrepareErrorJSON("Error", "Player doesn't have enough points", 400)
            }
            player.Points -= points
        case "fund":
            player.Points += points
    }
    c.Txn.Save(&player)
    return player
}

func (c BasicController) PreparePlayer(player models.Player) interface{} {
    data := make(map[string]interface{})
    data["playerId"] = player.ID
    data["balance"] = player.Points
    return data
}

func (c BasicController) IsPlayer(player interface{}) bool {
    _, ok := player.(models.Player)
    return ok
}

func (c BasicController) RenderPlayer(player interface{}) revel.Result {
    var data interface{}
    if c.IsPlayer(player) {
        data = c.PreparePlayer(player.(models.Player))
    } else {
        data = player
    }
    return c.RenderJSON(data)
}
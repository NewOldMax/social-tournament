package controllers

import (
    "strconv"
    "github.com/revel/revel"
    "app/app/models"
)

type Player struct {
    BasicController
}

// Route. Takes points from player account
func (c Player) Take() revel.Result {
    return c.ManagePoints("take")
    
}

// Route. Shows player balance
func (c Player) Balance() revel.Result {
    player := c.GetOrNotFoundPlayer()
    return c.RenderPlayer(player)
}

// Route. Adds points to balance. If no player exist should create new player with given amount of points
func (c Player) Fund() revel.Result {
    return c.ManagePoints("fund")
}

func (c Player) ManagePoints(action string) revel.Result {
    points, err :=  strconv.ParseFloat(c.Params.Query.Get("points"), 64)
    if err != nil {
        return c.RenderErrorJSON("Cannot parse points", 400)
    }
    var player interface{}
    if action == "take" {
        player = c.GetOrNotFoundPlayer()
    } else if action == "fund" {
        player = c.GetOrCreatePlayer()
    }
    if c.IsPlayer(player) {
        player = c.ManagePlayerPoints(player.(models.Player), points, action)
    }
    return c.RenderPlayer(player)
}
package controllers

import (
    "encoding/json"
    "strconv"
    "github.com/revel/revel"
    "app/app/models"
)

type Tournament struct {
    BasicController
}

// Route. Creates new tournament
func (c Tournament) Announce() revel.Result {
    tournamentId := c.Params.Query.Get("tournamentId")
    if tournamentId == "" {
        return c.RenderErrorJSON("Cannot parse tournamentId", 400)
    }
    var item models.Tournament
    c.Txn.Find(&models.Tournament{ID: tournamentId}).First(&item)
    if item.ID != "" {
        return c.RenderErrorJSON("Tournament with id '"+tournamentId+"' already exists", 400)
    }
    deposit, err := strconv.ParseFloat(c.Params.Query.Get("deposit"), 64)
    if err != nil {
        return c.RenderErrorJSON("Cannot parse deposit", 400)
    }
    if (deposit < 0) {
        return c.RenderErrorJSON("Deposit should be positive", 400)
    }
    var tournament models.Tournament
    tournament = models.Tournament{ID: tournamentId, Deposit: deposit, Status: "open", Prize: 0}
    c.Txn.Create(&tournament)
    return c.RenderCreatedTournament(tournament)
}

// Route. Join player and backers to tournament
func (c Tournament) Join() revel.Result {
    player := c.GetOrNotFoundPlayer()
    if c.IsPlayer(player) == false {
        return c.RenderJSON(player)
    }
    tournament := c.GetOpenedOrNotFoundTournament("GET")
    if c.IsTournament(tournament) == false {
        return c.RenderJSON(tournament)
    }
    if c.IsPlayerAlreadyJoined(player.(models.Player), tournament.(models.Tournament)) == true {
        return c.RenderErrorJSON("Player already joined", 400)
    }
    backers := c.GetBackers(player)
    _, ok := backers.([]models.Player)
    if ok == false {
        return c.RenderJSON(backers)
    }
    count := len(backers.([]models.Player)) + 1
    requiredAmount := tournament.(models.Tournament).Deposit / float64(count)
    if c.CheckAmount(requiredAmount, player.(models.Player), backers.([]models.Player)) == false {
        return c.RenderErrorJSON("Now enough points to join. Required "+strconv.FormatFloat(requiredAmount, 'f', -1, 64)+" from each player", 400)
    }
    c.JoinPlayersToTournament(player.(models.Player), backers.([]models.Player), tournament.(models.Tournament), requiredAmount)

    return c.RenderMessage("success")
}

// Route. Join player and backers to tournament
func (c Tournament) Result() revel.Result {
    tournament := c.GetOpenedOrNotFoundTournament("JSON")
    if c.IsTournament(tournament) == false {
        return c.RenderJSON(tournament)
    }
    winners := c.GetWinners()
    if c.IsWinnersBelongsToTournament(winners, tournament.(models.Tournament)) == false {
        return c.RenderErrorJSON("Not all winers are in tournament", 400)
    }
    if c.IsTournamentPrizeEqualToWinners(winners, tournament.(models.Tournament)) == false {
        return c.RenderErrorJSON("Winners prize doesn't match tournament prize", 400)
    }
    tournament = c.CloseTournament(winners, tournament.(models.Tournament))
    return c.RenderClosedTournament(tournament.(models.Tournament))
}

func (c Tournament) RenderClosedTournament(tournament models.Tournament) revel.Result {
    data := make(map[string]interface{})
    var winners interface{}
    json.Unmarshal([]byte(*tournament.Winners), &winners)
    data["winners"] = winners
    return c.RenderJSON(data)
}

func (c Tournament) CloseTournament(winners map[string]interface{}, tournament models.Tournament) models.Tournament {
    for _, winner := range winners {
        item := winner.(map[string]interface{})
        c.Params.Query.Set("playerId", item["playerId"].(string))
        player := c.GetOrNotFoundPlayer()
        if c.IsPlayer(player) {
            playerItem := player.(models.Player)
            c.GivePointsToWinners(playerItem, tournament, item["prize"].(float64))
        }
    }
    tournament.Status = "closed"
    preparedWinners := c.PrepareWinners(winners)
    tournament.Winners = &preparedWinners
    c.Txn.Save(&tournament)
    return tournament
}

func (c Tournament) GivePointsToWinners(player models.Player, tournament models.Tournament, prize float64) {
    var backers []models.TournamentPlayer
    c.Txn.Where(&models.TournamentPlayer{TournamentID: tournament.ID, Role: "backer", BackerForID: &player.ID}).Find(&backers)
    count := len(backers) + 1
    amount := prize / float64(count)
    if len(backers) > 0 {
        for _, backer := range backers {
            c.Params.Query.Set("playerId", backer.PlayerID)
            item := c.GetOrNotFoundPlayer().(models.Player)
            item.Points += amount
            c.Txn.Save(&item)
        }
    }
    player.Points += amount
    c.Txn.Save(&player)
    return
}

func (c Tournament) PrepareWinners(winners map[string]interface{}) string {
    var result []interface{}
    for _, winner := range winners {
        result = append(result, winner)
    }
    out, err := json.Marshal(result)
    if err != nil {
        panic(err)
    }
    return string(out)
}

func (c Tournament) IsTournamentPrizeEqualToWinners(winners map[string]interface{}, tournament models.Tournament) bool {
    prize := tournament.Prize
    calculatedPrize := 0.0
    for _, winner := range winners {
        item := winner.(map[string]interface{})
        calculatedPrize += item["prize"].(float64)
    }
    return calculatedPrize == prize
}

func (c Tournament) IsWinnersBelongsToTournament(winners map[string]interface{}, tournament models.Tournament) bool {
    result := true
    for _, winner := range winners {
        item := winner.(map[string]interface{})
        c.Params.Query.Set("playerId", item["playerId"].(string))
        player := c.GetOrNotFoundPlayer()
        if c.IsPlayer(player) {
            if c.IsPlayerAlreadyJoined(player.(models.Player), tournament) == false {
                result = false
            }
        }
        if result == false {
            return result
        }
    }
    return result
}

func (c Tournament) RenderCreatedTournament(tournament models.Tournament) revel.Result {
    c.Response.Status = 201
    data := make(map[string]interface{})
    data["tournamentId"] = tournament.ID
    data["deposit"] = tournament.Deposit
    data["status"] = tournament.Status
    return c.RenderJSON(data)
}

func (c Tournament) GetOpenedOrNotFoundTournament(method string) interface{} {
    var tournamentId string
    var err bool
    if method == "JSON" {
        var jsonData map[string]interface{}
        c.Params.BindJSON(&jsonData)
        tournamentId, err = jsonData["tournamentId"].(string)
        if err == false {
            tournamentId = ""
        }
    } else {
        tournamentId = c.Params.Query.Get("tournamentId")
    }
    if (tournamentId == "") {
        return c.PrepareErrorJSON("Error", "Cannot parse tournamentId", 400)
    }
    var tournament models.Tournament
    c.Txn.Where(&models.Tournament{ID: tournamentId, Status: "open"}).First(&tournament)
    if tournament.ID == "" {
        return c.PrepareErrorJSON("Not found", "Tournament with id '"+tournamentId+"' not found or closed", 404)
    }
    return tournament   
}

func (c Tournament) GetWinners() map[string]interface{} {
    var jsonData map[string]interface{}
    var uniqueWinners map[string]interface{}
    c.Params.BindJSON(&jsonData)
    winners := jsonData["winners"].([]interface{})
    if len(winners) > 0 {
        uniqueWinners = c.GetUniqueWinners(winners)
    }
    return uniqueWinners
}

func (c Tournament) GetUniqueWinners(winners []interface{}) map[string]interface{} {
    var tmp map[string]interface{}
    result := make(map[string]interface{})
    for _, element := range winners {
        item := element.(map[string]interface{})
        playerId := item["playerId"].(string)
        if playerId != "" {
            c.Params.Query.Set("playerId", playerId)
            player := c.GetOrNotFoundPlayer()
            if c.IsPlayer(player) {
                if result[playerId] == nil {
                    result[playerId] = make(map[string]interface{})
                    tmp = result[playerId].(map[string]interface{})
                    tmp["prize"] = 0.0
                    tmp["playerId"] = playerId
                    result[playerId] = tmp
                }
                tmp = result[playerId].(map[string]interface{})
                tmp["prize"] = tmp["prize"].(float64) + item["prize"].(float64)
                result[playerId] = tmp
            }
        }
    }
    return result
}

func (c Tournament) GetBackers(player interface{}) interface{} {
    var backers []models.Player
    item := c.Params.Query
    if len(item) > 0 {
        for _, element := range item["backerId"] {
            c.Params.Query.Set("playerId", element)
            backer := c.GetOrNotFoundPlayer()
            if c.IsPlayer(backer) {
                backer = backer.(models.Player)
                if (backer.(models.Player).ID == player.(models.Player).ID) {
                    return c.PrepareErrorJSON("Error", "Player can't be a backer", 400)
                }
                backers = append(backers, backer.(models.Player))
            } else {
                return backer
            }
        }
    }
    return backers
}

func (c Tournament) IsTournament(tournament interface{}) bool {
    _, ok := tournament.(models.Tournament)
    return ok
}

func (c Tournament) IsPlayerAlreadyJoined(player models.Player, tournament models.Tournament) bool {
    var entity models.TournamentPlayer
    c.Txn.Where(&models.TournamentPlayer{TournamentID: tournament.ID, PlayerID: player.ID, Role: "player"}).First(&entity)
    if entity.ID == 0 {
        return false
    }
    return true
}

func (c Tournament) CheckAmount(requiredAmount float64, player models.Player, backers []models.Player) bool {
    if player.Points < requiredAmount {
        return false
    }
    if len(backers) > 0 {
        for _, backer := range backers {
            if (backer.Points < requiredAmount) {
                return false
            }
        }
    }
    return true
}

func (c Tournament) JoinPlayersToTournament(player models.Player, backers []models.Player, tournament models.Tournament, requiredAmount float64) interface{} {
    prize := 0.0
    player = c.ManagePlayerPoints(player, requiredAmount, "take").(models.Player)
    if c.IsPlayer(player) == false {
        return player
    }
    prize += requiredAmount
    c.AddPlayerToTournament(player, tournament, requiredAmount, "player", player)
    if len(backers) > 0 {
        for _, backer := range backers {
            backer = c.ManagePlayerPoints(backer, requiredAmount, "take").(models.Player)
            if c.IsPlayer(backer) == false {
                return backer
            }
            prize += requiredAmount
            c.AddPlayerToTournament(backer, tournament, requiredAmount, "backer", player)
        }
    }
    tournament.Prize += prize
    c.Txn.Save(&tournament)
    return player
}

func (c Tournament) AddPlayerToTournament(player models.Player, tournament models.Tournament, points float64, role string, mainPlayer models.Player) interface{} {
    var tournamentPlayer models.TournamentPlayer
    tournamentPlayer = models.TournamentPlayer{TournamentID: tournament.ID, PlayerID: player.ID, Value: points, Role: role}
    c.Txn.Create(&tournamentPlayer)
    if role == "backer" {
        tournamentPlayer.BackerForID = &mainPlayer.ID
        c.Txn.Save(&tournamentPlayer)
    }
    return tournamentPlayer
}

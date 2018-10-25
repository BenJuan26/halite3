package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/BenJuan26/hlt"
	"github.com/BenJuan26/hlt/gameconfig"
	"github.com/BenJuan26/hlt/log"
)

func gracefulExit(logger *log.FileLogger) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig)
		fmt.Println("Wait for 2 second to finish processing")
		time.Sleep(2 * time.Second)
		logger.Close()
		os.Exit(0)
	}()
}

const (
	exploring = 1
	returning = 2
)

func main() {
	args := os.Args
	var seed = time.Now().UnixNano() % int64(os.Getpid())
	if len(args) > 1 {
		seed, _ = strconv.ParseInt(args[0], 10, 64)
	}
	rand.Seed(seed)

	var game = hlt.NewGame()
	// At this point "game" variable is populated with initial map data.
	// This is a good place to do computationally expensive start-up pre-processing.
	// As soon as you call "ready" function below, the 2 second per turn timer will start.

	var config = gameconfig.GetInstance()
	fileLogger := log.NewFileLogger(game.Me.ID)
	var logger = fileLogger.Logger
	logger.Printf("Successfully created bot! My Player ID is %d. Bot rng seed is %d.", game.Me.ID, seed)
	gracefulExit(fileLogger)
	var maxHalite, _ = config.GetInt(gameconfig.MaxHalite)
	var shipCost, _ = config.GetInt(gameconfig.ShipCost)
	maxTurns, _ := config.GetInt(gameconfig.MaxTurns)
	game.Ready("MyBot")
	shipRoles := make(map[int]int)
	shipTargets := make(map[int]*hlt.Position)
	for {
		game.UpdateFrame()
		var me = game.Me
		var gameMap = game.Map
		var ships = me.Ships
		var commands = []hlt.Command{}
		bank := 1000 * int(game.TurnNumber/20)

		for i := range ships {
			var ship = ships[i]
			shipID := ship.GetID()
			cellHalite := gameMap.AtEntity(ship.E).Halite
			if target, ok := shipTargets[shipID]; ok {
				logger.Printf("Ship %d: My target is %d,%d", shipID, target.GetX(), target.GetY())
			} else {
				logger.Printf("Ship %d: I still don't have a target", shipID)
			}

			if _, ok := shipRoles[shipID]; !ok {
				shipRoles[shipID] = exploring
			}

			if shipRoles[shipID] == returning {
				if ship.E.Pos.Equals(me.Shipyard.E.Pos) {
					shipRoles[shipID] = exploring
					cells := gameMap.CellsByHalite(me.Shipyard.E.Pos, -1)
					cell := cells[rand.Intn(4)]
					shipTargets[shipID] = cell.Pos
					commands = append(commands, ship.Move(gameMap.NaiveNavigate(ship, cell.Pos)))
					logger.Printf("Ship %d: Returned to base; switching to explore role", ship.GetID())
				} else {
					dir := gameMap.NaiveNavigate(ship, me.Shipyard.E.Pos)
					// just try a random position instead of standing still
					if dir.Equals(hlt.Still()) {
						perm := rand.Perm(5)
						for _, i := range perm {
							newDir := hlt.AllDirections[i]
							newPos, _ := ship.E.Pos.DirectionalOffset(newDir)
							normalized := gameMap.Normalize(newPos)
							if !newDir.Equals(hlt.Still()) && !gameMap.AtPosition(normalized).IsOccupied() {
								dir = newDir
								gameMap.AtPosition(normalized).MarkUnsafe(ship)
								logger.Printf("Ship %d: Navigation wanted me to stay, but I'm going %s instead", shipID, string(dir.GetCharValue()))
								break
							}
						}
					}
					commands = append(commands, ship.Move(dir))
				}
			} else if ship.Halite > (maxHalite / 2) {
				shipRoles[shipID] = returning
				if _, hasTarget := shipTargets[shipID]; hasTarget {
					delete(shipTargets, shipID)
				}
				logger.Printf("Ship %d: Halite is now at %d; returning to base", ship.GetID(), ship.Halite)
				commands = append(commands, ship.Move(gameMap.NaiveNavigate(ship, me.Shipyard.E.Pos)))
			} else if cellHalite < (maxHalite/10) && ship.Halite >= cellHalite/10 {
				if _, ok := shipTargets[shipID]; !ok {
					cells := gameMap.CellsByHalite(me.Shipyard.E.Pos, -1)
					cell := cells[rand.Intn(4)]
					shipTargets[shipID] = cell.Pos
				}
				dir := gameMap.NaiveNavigate(ship, shipTargets[shipID])
				// just try a random position instead of standing still
				if dir.Equals(hlt.Still()) {
					perm := rand.Perm(5)
					for _, i := range perm {
						newDir := hlt.AllDirections[i]
						newPos, _ := ship.E.Pos.DirectionalOffset(newDir)
						normalized := gameMap.Normalize(newPos)
						if !newDir.Equals(hlt.Still()) && !gameMap.AtPosition(normalized).IsOccupied() {
							dir = newDir
							gameMap.AtPosition(normalized).MarkUnsafe(ship)
							logger.Printf("Ship %d: Navigation wanted me to stay, but I'm going %s instead", shipID, string(dir.GetCharValue()))
							break
						}
					}
				}
				commands = append(commands, ship.Move(dir))
			} else {
				commands = append(commands, ship.Move(hlt.Still()))
				logger.Printf("Ship %d: Got to end of if block, staying still", ship.GetID())
			}
		}

		if game.TurnNumber <= maxTurns/2 && me.Halite >= shipCost+bank && !gameMap.AtEntity(me.Shipyard.E).IsOccupied() {
			commands = append(commands, hlt.SpawnShip{})
		}
		game.EndTurn(commands)
	}
}

func cellsByHalite(game *hlt.Game) []*hlt.MapCell {
	gm := game.Map
	me := game.Me
	maxHalite, _ := gameconfig.GetInstance().GetInt(gameconfig.MaxHalite)
	maxCellHalite := 0
	radius := 8
	var cells []*hlt.MapCell

	// We want at least 2 cells to have more than half halite
	for radius <= gm.GetWidth() && maxCellHalite < maxHalite/2 {
		cells := gm.CellsByHalite(me.Shipyard.E.Pos, 8)
		maxCellHalite = cells[3].Halite
	}
	return cells
}

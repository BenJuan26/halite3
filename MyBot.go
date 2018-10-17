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
	game.Ready("MyBot")
	shipRoles := make(map[int]int)
	for {
		game.UpdateFrame()
		var me = game.Me
		var gameMap = game.Map
		var ships = me.Ships
		var commands = []hlt.Command{}
		bank := 1000 * int(game.TurnNumber/40)

		for i := range ships {
			var ship = ships[i]
			shipID := ship.GetID()

			if _, ok := shipRoles[shipID]; !ok {
				shipRoles[shipID] = exploring
			}

			if shipRoles[shipID] == returning {
				if ship.E.Pos.Equals(me.Shipyard.E.Pos) {
					shipRoles[shipID] = exploring
					logger.Printf("Ship %d: Returned to base; switching to explore role", ship.GetID())
				} else {
					dir := gameMap.NaiveNavigate(ship, me.Shipyard.E.Pos)
					commands = append(commands, ship.Move(dir))
				}
			} else if ship.Halite > (maxHalite / 2) {
				shipRoles[shipID] = returning
				logger.Printf("Ship %d: Halite is now at %d; returning to base", ship.GetID(), ship.Halite)
			} else if gameMap.AtEntity(ship.E).Halite < (maxHalite / 10) {
				potentialDirection := hlt.AllDirections[rand.Intn(4)]
				newPos, _ := ship.E.Pos.DirectionalOffset(potentialDirection)
				if !gameMap.AtPosition(newPos).IsOccupied() {
					gameMap.AtPosition(newPos).MarkUnsafe(ship)
					commands = append(commands, ship.Move(potentialDirection))
				}
			} else {
				commands = append(commands, ship.Move(hlt.Still()))
				logger.Printf("Ship %d: Got to end of if block, staying still", ship.GetID())
			}
		}

		if me.Halite >= shipCost+bank && !gameMap.AtEntity(me.Shipyard.E).IsOccupied() {
			commands = append(commands, hlt.SpawnShip{})
		}
		game.EndTurn(commands)
	}
}

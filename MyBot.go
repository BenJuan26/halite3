package main

import (
	"fmt"
	"hlt"
	"hlt/gameconfig"
	"hlt/log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
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
	game.Ready("MyBot")
	for {
		game.UpdateFrame()
		var me = game.Me
		var gameMap = game.Map
		var ships = me.Ships
		var commands = []hlt.Command{}
		for i := range ships {
			var ship = ships[i]

			var currentCell = gameMap.AtEntity(ship.E)
			if ship.IsFull() {
				dir := gameMap.NaiveNavigate(ship, me.Shipyard.E.Pos)
				commands = append(commands, ship.Move(dir))
			} else if currentCell.Halite < (maxHalite / 10) {
				potentialDirection := hlt.AllDirections[rand.Intn(4)]
				potentialPos, _ := ship.E.Pos.DirectionalOffset(potentialDirection)
				if gameMap.AtPosition(potentialPos).IsEmpty() {
					commands = append(commands, ship.Move(potentialDirection))
				}
			} else {
				commands = append(commands, ship.Move(hlt.Still()))
			}
		}
		var shipCost, _ = config.GetInt(gameconfig.ShipCost)
		if game.TurnNumber <= 200 && me.Halite >= shipCost && !gameMap.AtEntity(me.Shipyard.E).IsOccupied() {
			commands = append(commands, hlt.SpawnShip{})
		}
		game.EndTurn(commands)
	}
}

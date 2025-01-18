package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	pubsub "github.com/bootdotdev/learn-pub-sub-starter/internal"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func HandlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {

	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")

		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func HandlerMove(gs *gamelogic.GameState, pubch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {

	return func(am gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		mv := gs.HandleMove(am)
		switch mv {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(pubch, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.GetUsername(), gamelogic.RecognitionOfWar{
				Attacker: am.Player,
				Defender: gs.GetPlayerSnap(),
			})
			if err != nil {
				fmt.Printf("error %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		fmt.Println("error: unknown move")
		return pubsub.NackDiscard
	}
}

func HandlerWar(gs *gamelogic.GameState, pubch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {

	return func(row gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")

		mv, winner, loser := gs.HandleWar(row)

		switch mv {
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeDraw:
			err := pubsub.PublishGob(pubch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+gs.GetUsername(), routing.GameLog{
				CurrentTime: time.Now(),
				Username:    gs.GetUsername(),
				Message:     fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser),
			})
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeOpponentWon:
			err := pubsub.PublishGob(pubch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+gs.GetUsername(), routing.GameLog{
				CurrentTime: time.Now(),
				Username:    gs.GetUsername(),
				Message:     fmt.Sprintf("%s won a war against %s", winner, loser),
			})
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			err := pubsub.PublishGob(pubch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+gs.GetUsername(), routing.GameLog{
				CurrentTime: time.Now(),
				Username:    gs.GetUsername(),
				Message:     fmt.Sprintf("%s won a war against %s", winner, loser),
			})
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}

		fmt.Printf("error reading the war outcome")
		return pubsub.NackDiscard
	}
}

func main() {
	fmt.Println("Starting Peril client...")
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	pubchannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel %v", err)
	}
	gamestate := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+gamestate.GetUsername(), routing.ArmyMovesPrefix+".*", pubsub.TransientQueue, HandlerMove(gamestate, pubchannel))
	if err != nil {
		fmt.Printf("could not subscribe to army moves %v", err)
	}

	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".*", pubsub.DurableQueue, HandlerWar(gamestate, pubchannel))
	if err != nil {
		fmt.Printf("could not subscribe to war %v", err)
	}

	pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, "pause."+gamestate.GetUsername(), routing.PauseKey, pubsub.TransientQueue, HandlerPause(gamestate))
	if err != nil {
		fmt.Printf("could not subscribe to pause %v", err)
	}

	for {
		command := gamelogic.GetInput()
		if len(command) == 0 {
			continue
		}

		switch command[0] {

		case "spawn":
			err = gamestate.CommandSpawn(command)
			if err != nil {
				log.Println(err)
				continue
			}
		case "move":
			am, err := gamestate.CommandMove(command)
			if err != nil {
				log.Println(err)
				continue
			}
			err = pubsub.PublishJSON(pubchannel, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+am.Player.Username, am)

			if err != nil {
				log.Printf("err %s\n", err)
			}
			fmt.Printf("Moved %v units to %s\n", len(am.Units), am.ToLocation)
		case "status":
			gamestate.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			if len(command) != 2 {
				gamelogic.PrintClientHelp()
				return
			}
			n, err := strconv.Atoi(command[1])
			if err != nil {
				log.Println("cannot convert string value to integer")
			}
			for i := 0; i < n; i++ {
				spam := gamelogic.GetMaliciousLog()
				err=pub
			}
			log.Println("spamming not allowed")

		case "quit":
			log.Println("exiting")
			return

		default:
			log.Println("unknown command")
			gamelogic.PrintClientHelp()
		}
	}

}

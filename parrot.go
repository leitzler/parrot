package main

import (
	"log"
	"os"

	"github.com/leitzler/parrot/bot"
	"github.com/nlopes/slack"
)

func main() {
	slackToken := os.Getenv("PARROT_SLACK_TOKEN")
	if slackToken == "" {
		log.Fatal("No token set, use environment var PARROT_SLACK_TOKEN")
	}
	api := slack.New(slackToken)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	b := bot.New(api)

	go func() {
		for msg := range rtm.IncomingEvents {
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				log.Println("Hello event received, we are connected!")
			case *slack.ConnectingEvent:
				log.Println("Connecting..")
			case *slack.MessageEvent:
				go b.HandleMessage(ev)
			case *slack.InvalidAuthEvent:
				log.Fatalln("Invalid auth. Check your $PARROT_SLACK_TOKEN.")
			}
		}
	}()

	select {}
}

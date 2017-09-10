package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nlopes/slack"
	"github.com/saniales/golang-crypto-trading-bot/cmd"
	"github.com/saniales/golang-crypto-trading-bot/environment"
	"github.com/saniales/golang-crypto-trading-bot/exchangeWrappers"
	"github.com/saniales/golang-crypto-trading-bot/strategies"
	"github.com/shomali11/slacker"
)

var slackBot = slacker.NewClient(os.Getenv("SLACK_TOKEN"))
var strategy = strategies.IntervalStrategy{
	Model: strategies.StrategyModel{
		Name: "RossBotuccino",
		Setup: func(exchangeWrappers.ExchangeWrapper, *environment.Market) error {
			// connect slack token
			slackBot.Init(func() {
				log.Println("Slackbot Connected")
			})
			slackBot.Err(func(err string) {
				log.Println("Error during slack bot connection: ", err)
			})
			go func() {
				err := slackBot.Listen()
				if err != nil {
					log.Fatal(err)
				}
			}()
			return nil
		},
		OnUpdate: func(ew exchangeWrappers.ExchangeWrapper, m *environment.Market) error {
			channelID, timestamp, err := slackBot.Client.PostMessage(os.Getenv("POST_CHANNEL"), "OMG something happening!!!!! "+m.Name+": "+fmt.Sprint(m.Summary.Last), slack.PostMessageParameters{AsUser: true})
			log.Printf("Message Posted, CHANNEL ID: %s, TIMESTAMP: %s\n", channelID, timestamp)
			return err
		},
		OnError: func(err error) {
			slackBot.Client.PostMessage(os.Getenv("POST_CHANNEL"), "I am dead due to an error @thebotguy; "+err.Error(), slack.PostMessageParameters{})
		},
	},
	Interval: time.Second * 10,
}

func main() {
	strategies.AddCustomStrategy(strategy)
	bot.Execute()
}

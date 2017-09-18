package main

import (
	"log"
	"os"
	"time"

	"github.com/markcheno/go-talib"

	"github.com/saniales/golang-crypto-trading-bot/cmd"
	"github.com/saniales/golang-crypto-trading-bot/environment"
	"github.com/saniales/golang-crypto-trading-bot/exchangeWrappers"
	"github.com/saniales/golang-crypto-trading-bot/strategies"
	"github.com/shomali11/slacker"
	"github.com/thebotguys/golang-bittrex-api/bittrex"
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
			candles, err := bittrex.GetTicks(m.Name, "thirtyMin")
			if err != nil {
				return err
			}
			closePrices := make([]float64, len(candles))
			for i, candle := range candles {
				closePrices[i] = candle.Close
			}

			RSI30M := talib.Rsi(closePrices, 14)[14]
			candles, err = bittrex.GetTicks(m.Name, "hour")
			if err != nil {
				return err
			}
			closePrices = make([]float64, len(candles))
			for i, candle := range candles {
				closePrices[i] = candle.Close
			}

			RSI1H := talib.Rsi(closePrices, 14)[14]

			MACD, _, _ := talib.Macd(closePrices, 12, 26, 9)
			log.Println("MARKET:", m.Name)
			log.Println("RSI (30 MIN):", RSI30M)
			log.Println("RSI (1 HOUR):", RSI1H)
			log.Println("MACD: ", MACD[9-1+26-1])
			//log.Println(MACD[9 - 1 + 26 - 1])
			//log.Println(Signal[12])
			//log.Println(low[32])
			/*
				channelID, timestamp, err := slackBot.Client.PostMessage(
					os.Getenv("POST_CHANNEL"),
					fmt.Sprintln("TEST SIGNAL ON ", m.Name, " MARKET:")+
						fmt.Sprintln("RSI (30 Min): ", RSI30M)+
						fmt.Sprintln("RSI (1 Hour): ", RSI1H),
					//	fmt.Sprintln("RSI (4 Hours): ", RSI4H)
					//fmt.Sprintln("MACD: { \n    MACD:", MACD[32], "\n    SIGNAL: ", signal[32], "\n    LOW: ", low[32], "\n}"),
					slack.PostMessageParameters{AsUser: true})
				log.Printf("Message Posted, CHANNEL ID: %s, TIMESTAMP: %s\n", channelID, timestamp)
			*/
			return nil
		},
		OnError: func(err error) {
			//slackBot.Client.PostMessage(os.Getenv("POST_CHANNEL"), "I am dead due to an error @thebotguy; "+err.Error(), slack.PostMessageParameters{})
		},
	},
	Interval: time.Minute * 5,
}

func main() {
	strategies.AddCustomStrategy(strategy)
	bot.Execute()
}

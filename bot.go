package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/markcheno/go-talib"
	"github.com/nlopes/slack"

	"github.com/saniales/golang-crypto-trading-bot/cmd"
	"github.com/saniales/golang-crypto-trading-bot/environment"
	"github.com/saniales/golang-crypto-trading-bot/exchangeWrappers"
	"github.com/saniales/golang-crypto-trading-bot/strategies"
	"github.com/shomali11/slacker"
	"github.com/thebotguys/golang-bittrex-api/bittrex"
)

var oldSma20 []float64
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
			err := ew.GetMarketSummary(m)
			if err != nil {
				return err
			}
			candles, err := bittrex.GetTicks(m.Name, "thirtyMin")
			if err != nil {
				return err
			}
			closePrices := make([]float64, len(candles))
			for i, candle := range candles {
				closePrices[i] = candle.Close
				//fmt.Println(time.Time(candle.Timestamp)) // ok, sorted
			}

			sma20 := talib.Sma(closePrices, 20)
			sma50 := talib.Sma(closePrices, 50)

			if oldSma20 != nil {
				ok := false
				for i := len(oldSma20) - 1; i > len(sma20)-12; i-- {
					if sma20[len(sma20)-i] != oldSma20[i] {
						ok = true
						break
					}
				}
				//log.Println("OLD:", oldSma20)
				//log.Println("NEW:", sma20[len(sma20)-11:len(sma20)-1])
				if !ok {
					log.Println("SKIPPED, EQUAL")
					return nil
				}
			}
			oldSma20 = append([]float64(nil), sma20[len(sma20)-11:len(sma20)-1]...)
			var message string
			i := len(sma20) - 1
			if sma20[i] > sma50[i] && sma20[i-1] < sma50[i-1] { //sell
				log.Println("BUY CONDITION MET AT ", m.Summary.Ask)
				message = fmt.Sprintf(
					`BUY SIGNAL ON %s:
SMA(20) CROSSING UP SMA(50)
BUY AT %f`, m.Name, m.Summary.Ask)
			} else if sma20[i] < sma50[i] && sma20[i-1] > sma50[i-1] { //buy
				log.Println("SELL CONDITION MET AT ", m.Summary.Bid)
				message = fmt.Sprintf(
					`SELL SIGNAL ON %s:
SMA(20) CROSSING DOWN SMA(50)
SELL AT %f`, m.Name, m.Summary.Bid)
			} else {
				log.Println("NO ACTION")
			}

			log.Println("SMA(20): ", sma20[len(sma20)-1], sma20[len(sma20)-2])
			log.Println("SMA(50): ", sma50[len(sma50)-1], sma50[len(sma50)-2])

			if message == "" {
				return nil
			}

			channelID, timestamp, err := slackBot.Client.PostMessage(
				os.Getenv("POST_CHANNEL"), message,
				slack.PostMessageParameters{AsUser: true})
			log.Printf("Message Posted, CHANNEL ID: %s, TIMESTAMP: %s\n", channelID, timestamp)
			slackBot.Client.PostMessage(
				os.Getenv("POST_CHANNEL"), fmt.Sprintf("<@U6NKZ63KR> graph %s %s 24h", m.MarketCurrency, m.BaseCurrency),
				slack.PostMessageParameters{AsUser: true})
			return nil
		},
		OnError: func(err error) {
			slackBot.Client.PostMessage(os.Getenv("POST_CHANNEL"), "I am dead due to an error <@thebotguy>; "+err.Error(), slack.PostMessageParameters{AsUser: true})
		},
	},
	Interval: time.Minute * 5,
}

func main() {
	strategies.AddCustomStrategy(strategy)
	bot.Execute()
}

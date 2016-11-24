package main

import (
	"fmt"
	"log"
	"os"
	"flag"
	"github.com/yhpark/hangul-mealy"
	"github.com/nlopes/slack"
)

func main() {
	token := flag.String("t", "", "Slack API Token")

	flag.Parse()

	if *token == "" {
		fmt.Println("Run with `-h` for usage")
		return
	}

	api := slack.New(*token)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	enabled := make(map[string]bool)

loop:	for {
		select {
		case msg := <-rtm.IncomingEvents:
			fmt.Printf("Event Received: %T\n", msg.Data)
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				// ignore hello

			case *slack.MessageEvent:
				if ev.Text == "!kor" || ev.Text == "!korean" {
					enabled[ev.User] = !enabled[ev.User]
					if enabled[ev.User] {
						fmt.Printf("kor-typer enabled for %s\n", ev.User)
						rtm.SendMessage(rtm.NewOutgoingMessage("hello!", ev.Channel))
					} else {
						fmt.Printf("kor-typer disabled for %s\n", ev.User)
						rtm.SendMessage(rtm.NewOutgoingMessage("bye..", ev.Channel))
					}
					break
				}

				if !enabled[ev.User] {
					break
				}

				fmt.Printf("Message: %s\n", ev.Text)

				mealy, e := hangulmealy.MakeHangulMealy(false)
				if e != nil {
					fmt.Println(e.Error())
					break
				}

				e = mealy.RunEng(ev.Text)
				if e != nil {
					fmt.Println(e.Error())
					break
				}

				hangul := mealy.HangulString()
				fmt.Printf("h: %s, ch: %s\n", hangul, ev.Channel)

				rtm.SendMessage(rtm.NewOutgoingMessage(hangul, ev.Channel))

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break loop

			default:
				fmt.Printf("Ignored\n")
			}
		}
	}
}

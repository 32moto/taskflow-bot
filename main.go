package main

import (
	"os"
	"net/http"
	"fmt"
	"log"
	"io/ioutil"
	"encoding/json"
	"github.com/line/line-bot-sdk-go/linebot"
)

type Ping struct {
	Status int    `json:"status"`
	Ping   string `json:"ping"`
}

type PushMessageParams struct {
	To string `json:"to"`
	MessageType string `json:"message_type"`
	Message string `json:"message"`
}

func main() {
	bot, err := linebot.New(
		os.Getenv("LINE_BOT_CHANNEL_SECRET"),
		os.Getenv("LINE_BOT_CHANNEL_ACCESS_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/push_message", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			ping := Ping{http.StatusOK, "pong!"}

			res, err := json.Marshal(ping)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(res)
		case "POST":
			body, err := ioutil.ReadAll(req.Body)

			if err != nil {
				fmt.Println("io error")
				return
			}

			bodyJsonBytes := ([]byte)(body)

			data := new(PushMessageParams)
			if err := json.Unmarshal(bodyJsonBytes, data); err != nil {
				fmt.Println("JSON Unmarshal error:", err)
				return
			}

			switch data.MessageType {
			case "text":
				_, err := bot.PushMessage(
					data.To,
					linebot.NewTextMessage(data.Message),
				).Do()
				if err != nil {
					log.Fatal(err)
				}

			case "flexMessage":
				jsonData := []byte(data.Message)

				container, err := linebot.UnmarshalFlexMessageJSON(jsonData)
				if err != nil {
					fmt.Println("JSON Unmarshal error:", err)
					return
				}
				res, err := bot.PushMessage(
					data.To,
					linebot.NewFlexMessage("notify from Taskflow", container),
				).Do()
				if err != nil || res == nil {
					log.Fatal(err)
				}
			}
		}
	})

	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
				  if message.Text == "check" {
            ids := "【check ids】\n"
            ids += " UserId ----------\n" + event.Source.UserID + "\n\n"
            ids += " GroupId ----------\n" + event.Source.GroupID + "\n\n"
            ids += " RoomId ----------\n" + event.Source.RoomID

            if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(ids)).Do(); err != nil {
              log.Print(err)
            }
					}
				}
			}
		}
	})
	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}

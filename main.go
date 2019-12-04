package main

import (
	"github.com/Syfaro/telegram-bot-api"
	"io/ioutil"
	"log"
	"nanomsg.org/go/mangos/v2"
	_ "nanomsg.org/go/mangos/v2/protocol/rep"
	"nanomsg.org/go/mangos/v2/protocol/req"
	"net/http"

	// register transports
	_ "nanomsg.org/go/mangos/v2/transport/all"
)

var sock mangos.Socket

const Token string = "1004887434:AAEfxr8GZPsTqv1IlUuuzpb9h5iTo_Vd1cw"
const DaemonUrl string = "tcp://localhost:8000"

func DownloadFile(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	return body
}

func IsPizza(fileBytes []byte) bool {
	var err error
	var msg []byte

	if err = sock.Send(fileBytes); err != nil {
		log.Panic(err)
	}

	if msg, err = sock.Recv(); err != nil {
		log.Panic(err)
	}

	ret := string(msg[:len(msg)-1])

	log.Printf("Pizza daemon returned: %s\n", ret)

	return ret == "pizza"
}

func main() {
	var err error

	if sock, err = req.NewSocket(); err != nil {
		log.Panic(err)
	}
	defer sock.Close()

	log.Printf("Dialing to %s", DaemonUrl)
	if err = sock.Dial(DaemonUrl); err != nil {
		log.Panic(err)
	}
	log.Println("Success")

	log.Println("Connecting to a Telegram bot")
	bot, err := tgbotapi.NewBotAPI(Token)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Success")
	log.Printf("Authorized on account %s", bot.Self.UserName)

	bot.Debug = false

	var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60
	updates, err := bot.GetUpdatesChan(ucfg)

	for update := range updates {
		if update.Message == nil {
			log.Println("No message in update")
			continue
		}

		if err != nil {
			log.Panic(err)
		}

		chatId := update.Message.Chat.ID
		photos := *update.Message.Photo

		log.Printf("Message received from chat %d", chatId)

		fileConfig := tgbotapi.FileConfig{FileID: photos[0].FileID}

		log.Printf("File id received %s", fileConfig.FileID)

		file, err := bot.GetFile(fileConfig)

		if err != nil {
			log.Panic(err)
		}

		fileBytes := DownloadFile(file.Link(Token))

		var msg tgbotapi.MessageConfig
		if IsPizza(fileBytes) == true {
			msg = tgbotapi.NewMessage(chatId, "This is a pizza")
		} else {
			msg = tgbotapi.NewMessage(chatId, "This is not a pizza")
		}

		if _, err = bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

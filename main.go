package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type binanceResp struct {
	Price float64 `json: "price,string"`
	Code  int64   `json: "code"`
}

func getKey() string {
	return "2120064016:AAEt-sdqchuVskjVOHXSMk3tGhWTGHqvWWY"
}

type wallet map[string]float64

var db = map[int64]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI(getKey())
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		/*
			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello, "+update.CallbackQuery.From.FirstName)
				bot.Send(msg)
				continue
			}
		*/

		log.Println(update.Message.Text)

		msgArr := strings.Split(update.Message.Text, " ")
		switch msgArr[0] {
		case "ADD":
			sum, err := strconv.ParseFloat(msgArr[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Невозможно сконвертировать валюту"))
				continue
			}

			if _, ok := db[update.Message.Chat.ID]; !ok {
				db[update.Message.Chat.ID] = wallet{}
			}
			db[update.Message.Chat.ID][msgArr[1]] += sum

			msg := fmt.Sprintf("Balance: %s %f", msgArr[1], db[update.Message.Chat.ID][msgArr[1]])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))

		case "SUB":
			sum, err := strconv.ParseFloat(msgArr[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Невозможно сконвертировать валюту"))
				continue
			}

			if _, ok := db[update.Message.Chat.ID]; !ok {
				db[update.Message.Chat.ID] = wallet{}
			}
			db[update.Message.Chat.ID][msgArr[1]] -= sum

			msg := fmt.Sprintf("Balance: %s %f", msgArr[1], db[update.Message.Chat.ID][msgArr[1]])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		case "DEL":
			delete(db[update.Message.Chat.ID], msgArr[1])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Валюта удалена"))
		case "SHOW":
			msg := "Balance:\n"
			var usedSum float64
			for key, value := range db[update.Message.Chat.ID] {
				coinPrice, err := getPrice(key)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				}
				usedSum += value * coinPrice
				msg += fmt.Sprintf("%s: %.1f [%.1f]\n", key, value, value*coinPrice)
			}
			msg += fmt.Sprintf("Сумма: %.1f\n", usedSum)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда"))
		}
	}
}

func getPrice(coin string) (price float64, err error) {
	resp, err := http.Get(fmt.Sprintf("http://api.binance.com/api/v3/ticker/price?symbol=%sUAH", coin))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var jsonResp binanceResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}

	if jsonResp.Code != 0 {
		err = errors.New("Invalid valute")
		return
	}

	price = jsonResp.Price

	return
}

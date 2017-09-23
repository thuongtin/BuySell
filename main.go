package main

import (
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"gopkg.in/mgo.v2"
	"time"
	"sync"
	"strings"
	"strconv"
	"fmt"
)

type Person struct {
	Name string
	Phone string
}
type m bson.M
type Vol struct {
	Buy, Sell float64
}

type r struct {
	Time int64 `bson:"_id"`
	BuyVol float64 `bson:"buyVol"`
	SellVol float64 `bson:"sellVol"`
}

var session *mgo.Session
var mutex sync.Mutex
func main() {


	s, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer s.Close()
	bot, err := tgbotapi.NewBotAPI("400916444:AAGXPNPCrfJsIfxsUX0ryfEGi4H0SavEu_0")
	if err != nil {
		log.Panic(err)
	}
	session = s
	session.SetMode(mgo.Monotonic, true)
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		if !update.Message.IsCommand() {
			continue
		}

		marr := strings.Split(update.Message.Text, " ")
		str := ""
		if len(marr) >= 2 && marr[0] == "/vol" {
			minutes := 5
			if len(marr) == 3 {
				minutes, err = strconv.Atoi(marr[2])
				if err != nil {
					continue
				}
			}
			pair := marr[1]

			vols, e := getVol(pair, minutes * 60)
			if e != nil || len(vols) == 0 {
				str = "Không lấy được volume của cặp: " + pair
			} else {
				for _, vol := range vols {
					buyP, sellP := getPercent(vol.BuyVol, vol.SellVol)
					str += fmt.Sprintf("%s - Buy: %.3f (%.2f%%) - Sell %.3f (%.2f%%)\n", time.Unix(vol.Time/1000, 0).Format("2006-01-02 15:04"), vol.BuyVol, buyP, vol.SellVol, sellP)
				}
			}
			//update.Message.Chat.ID
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, str)
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}


	}
}

func getVol(pair string, seconds int) ([]r, error) {
	mutex.Lock()
	defer mutex.Unlock()
	c := session.DB("bittrex").C(pair)
	pipeline := []m{
		{
			"$group": m{
				"_id": m{
					"$subtract": []m{
						{ "$subtract": []interface{}{"$t", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)}},
						{"$mod":[]interface{}{m{"$subtract":[]interface{}{"$t", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)}},1000*seconds}}},

				},
				"sellVol": m{ "$sum": "$s" },
				"buyVol": m{ "$sum": "$b" },
			},
		},
		{
			"$sort": m{"_id": -1},
		},
		{
			"$limit": 10,
		},

	}
	pipe := c.Pipe(pipeline)
	resp := []r{}
	err := pipe.All(&resp)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%#v\n", resp)
	return resp, nil
}


func getPercent(buy, sell float64) (float64, float64){
	total := buy + sell
	onePercent := total/100
	return buy/onePercent , sell/onePercent
}

//func main() {
//	result := make(map[string]map[int64]Vol)
//	seconds := int64(300)
//	now := time.Now().UTC().Unix()*1000
//	_now := now - (now%(seconds*1000))
//	first := _now - (seconds*1000 * 24)
//	session, err := mgo.Dial("localhost")
//	if err != nil {
//		panic(err)
//	}
//	defer session.Close()
//	// Optional. Switch the session to a monotonic behavior.
//	session.SetMode(mgo.Monotonic, true)
//	names, _ := session.DB("bittrex").CollectionNames()
//	for _, name := range names {
//		result[name] = make(map[int64]Vol)
//		c := session.DB("bittrex").C(name)
//		pipeline := []m{
//			{
//				"$group": m{
//					"_id": m{
//						"$subtract": []m{
//							{ "$subtract": []interface{}{"$t", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)}},
//							{"$mod":[]interface{}{m{"$subtract":[]interface{}{"$t", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)}},1000*seconds}}},
//
//					},
//					"sellVol": m{ "$sum": "$s" },
//					"buyVol": m{ "$sum": "$b" },
//				},
//			},
//			{
//				"$match": m{"_id":m{"$gte":first}},
//			},
//			{
//				"$sort": m{"_id": -1},
//			},
//		}
////
//		pipe := c.Pipe(pipeline)
//		resp := []r{}
//		err := pipe.All(&resp)
//		if err != nil {
//			//handle error
//			panic(err)
//		}
//
//		for i:=first; i<=now; i+=seconds*100 {
//			result[name][i] = Vol{0,0}
//			for _, item := range resp {
//				if item.Time == i {
//					result[name][i] = Vol{item.BuyVol, item.SellVol}
//					break
//				}
//			}
//		}
//	}
//
//	for pair, item := range result {
//		fmt.Println(pair)
//		fmt.Printf("%#v\n\n", item[_now])
//	}
//}
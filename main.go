package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
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

func main() {
	result := make(map[string]map[int64]Vol)
	seconds := int64(1800)
	now := time.Now().UTC().Unix()*1000
	_now := now - (now%(seconds*1000))
	first := _now - (seconds*1000 * 24)
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	names, _ := session.DB("bittrex").CollectionNames()
	for _, name := range names {
		result[name] = make(map[int64]Vol)
		c := session.DB("bittrex").C(name)
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
				"$match": m{"_id":m{"$gte":first}},
			},
			{
				"$sort": m{"_id": -1},
			},
		}
//
		pipe := c.Pipe(pipeline)
		resp := []r{}
		err := pipe.All(&resp)
		if err != nil {
			//handle error
			panic(err)
		}

		for i:=first; i<=now; i+=seconds*100 {
			result[name][i] = Vol{0,0}
			for _, item := range resp {
				if item.Time == i {
					result[name][i] = Vol{item.BuyVol, item.SellVol}
					break
				}
			}
		}
	}

	for pair, item := range result {
		fmt.Println(pair)
		fmt.Printf("%#v\n\n", item[_now])
	}
}
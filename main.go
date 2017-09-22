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
func main() {
	now := time.Now().UTC().Unix()
	fmt.Println(now - (now%300))
	panic("")
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	names, _ := session.DB("bittrex").CollectionNames()
	for _, name := range names {
		c := session.DB("bittrex").C(name)
		pipeline := []m{
			{
				"$group": m{
					"_id": m{
						"$subtract": []m{
							{ "$subtract": []interface{}{"$t", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)}},
							{"$mod":[]interface{}{m{"$subtract":[]interface{}{"$t", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)}},1000*60*5}}},

					},
					"sellVol": m{ "$sum": "$s" },
					"buyVol": m{ "$sum": "$b" },
				},
			},
			{
				"$sort": m{"_id": -1},
			},
		}

		pipe := c.Pipe(pipeline)
		resp := []bson.M{}
		err := pipe.All(&resp)
		if err != nil {
			//handle error
		}
		fmt.Println(name)
		fmt.Printf("%#v", resp) // simple print proving it's working
		fmt.Println("\n\n\n")
	}
}
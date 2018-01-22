package main

import (
	"fmt"
	bw2 "github.com/gtfierro/bw2bind"
	"time"
)

func main() {
	c := bw2.ConnectOrExit("")
	c.SetEntityFromEnvironOrExit()
	c.OverrideAutoChainTo(true)

	i := 0
	for {
		time.Sleep(1 * time.Second)
		fmt.Println(i)
		i += 1
		msg := map[string]interface{}{
			"value": i,
			"time":  time.Now().UnixNano(),
		}
		po, err := bw2.CreateMsgPackPayloadObject(bw2.FromDotForm("2.0.0.0"), msg)
		if err != nil {
			fmt.Println("serialize", err)
			continue
		}
		err = c.Publish(&bw2.PublishParams{
			URI:            "scratch.ns/reliabletest",
			Persist:        true,
			PayloadObjects: []bw2.PayloadObject{po},
		})
		if err != nil {
			fmt.Println("publish", err)
		}
	}
}

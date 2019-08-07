package main

import (
	"BiliBan/src/BiliBan"
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:8080", nil))
	}()
	baseCtx := context.Background()
	checkCenter := &BiliBan.CheckCenter{}
	msgIn := checkCenter.Init(20, 10, BiliBan.FuncList{BiliBan.Filter_theSameCode}, BiliBan.FuncList{BiliBan.Filter_speed},
		map[string]interface{}{
			"BiliBan.Filter_theSameCode": float32(0.75),
			"BiliBan.Filter_speed": BiliBan.Filter_speed_config{
				StartCheck: 2,
				Limit:      float32(0.75),
			},
		})
	PopularList, err := BiliBan.GetPopular(100)
	if err != nil {
		log.Panic("丢失视野")
	}
	for _, roomId := range PopularList.Get("data.#.roomid").Array() {
		liveRoom := &BiliBan.LiveRoom{
			RoomID: roomId.Uint(),
			ReceiveMsg: func(model *BiliBan.MsgModel) {
				msgIn <- model
			},
		}
		go liveRoom.Start(baseCtx)
	}
	checkCenter.Start()
}

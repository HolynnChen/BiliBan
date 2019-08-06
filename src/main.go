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
	liveRoom := &BiliBan.LiveRoom{
		RoomID: 5050,
		ReceiveMsg: func(model *BiliBan.MsgModel) {
			msgIn <- model
		},
	}
	checkCenter.Start()
	liveRoom.Start(baseCtx)
}

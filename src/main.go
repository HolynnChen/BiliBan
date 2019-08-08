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
	msgIn := checkCenter.Init(20, 10, BiliBan.FuncList{BiliBan.Filter_theSameCode}, BiliBan.FuncList{BiliBan.Filter_speed, BiliBan.Filter_checkModels},
		&BiliBan.ConfigMap{
			Filter_theSameCode_limit:  float32(0.45),
			Filter_speed_StartCheck:   2,
			Filter_speed_Limit:        float32(0.75),
			Filter_checkModels_limit:  float32(0.75),
			Filter_checkModels_models: []string{"##o######汝逼q-", "####o#####逼q#", "##########曰汝σ#", "###oo####曰汝q", "########汝色σ", "寂寞十qq:##########", "########叭#姐姐", "######➏##幼籹", "########∃∃萝莉q", "э#########學嫂q", "#э########妹水+q"},
		})
	PopularList, err := BiliBan.GetPopular(300)
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

package main

import (
	"BiliBan/src/BiliBan"
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"regexp"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:8080", nil))
	}()
	baseCtx := context.Background()
	checkCenter := &BiliBan.CheckCenter{}
	reg1, _ := regexp.Compile(`\d`)
	reg2, _ := regexp.Compile(`[.|/\@~*&^ +-]`)
	msgIn := checkCenter.Init(20, 10, BiliBan.FuncList{BiliBan.Filter_theSameCode}, BiliBan.FuncList{BiliBan.Filter_checkRecent, BiliBan.Filter_speed, BiliBan.Filter_checkModels},
		&BiliBan.ConfigMap{
			Filter_theSameCode_limit:  float32(0.45),
			Filter_speed_StartCheck:   2,
			Filter_speed_Limit:        float32(0.75),
			Filter_checkModels_limit:  float32(0.75),
			Filter_checkModels_models: []string{},
			Filter_checkModels_expend: []*BiliBan.RegVal{&BiliBan.RegVal{Compiled: reg1, Value: "#"}, &BiliBan.RegVal{Compiled: reg2, Value: ""}},
			Filter_checkRecent_limit:  5,
		})
	//创建热门房间
	PopularList, err := BiliBan.GetPopular(500)
	if err != nil {
		log.Panic("丢失视野")
	}
	RoomRaw := PopularList.Get("data.#.roomid").Array()
	RoomList := BiliBan.AllToUnit(&RoomRaw)
	RoomCover := make(chan uint64, 10)
	for _, roomId := range *RoomList {
		enterRoom(&baseCtx, roomId, &msgIn, &RoomCover)
	}
	//房间切换
	go func() {
		for {
			select {
			case roomId := <-RoomCover:

				oldRooms := BiliBan.UnitToMap(BiliBan.AllToUnit(&RoomRaw))
				PopularList, err := BiliBan.GetPopular(500)
				if err != nil {
					log.Panic("丢失视野")
				}
				RoomRaw := PopularList.Get("data.#.roomid").Array()
				newRoomList := BiliBan.AllToUnit(&RoomRaw)
				for _, value := range *newRoomList {
					if value == roomId {
						continue
					}
					if _, exits := (*oldRooms)[value]; !exits {
						enterRoom(&baseCtx, value, &msgIn, &RoomCover)
						log.Printf("自动切换房间，原房间%d，当前房间%d", roomId, value)
						return
					}
				}
			}
		}
	}()
	//开始执行
	checkCenter.Start()
}

func enterRoom(baseCtx *context.Context, RoomID uint64, msgIn *chan *BiliBan.MsgModel, RoomCover *chan uint64) {
	liveRoom := &BiliBan.LiveRoom{
		RoomID: RoomID,
		ReceiveMsg: func(model *BiliBan.MsgModel) {
			*msgIn <- model
		},
		Preparing: func(RoomID uint64) {
			*RoomCover <- RoomID
		},
	}
	go liveRoom.Start(*baseCtx)
}

package main

import (
	"BiliBan/src/BiliBan"
	"context"
	"log"
	_ "net/http/pprof"
	"regexp"
	"time"
)

func main() {
	//go func() {
	//	log.Println(http.ListenAndServe("localhost:8080", nil))
	//}()
	baseCtx := context.Background()
	checkCenter := &BiliBan.CheckCenter{}
	roomMax := 800
	reg1, _ := regexp.Compile(`\d`)
	reg2, _ := regexp.Compile(`[.|/\@~*&^ +-]`)
	msgIn := checkCenter.Init(20, 10, BiliBan.FuncList{BiliBan.Filter_keyword, BiliBan.Filter_theSameCode}, BiliBan.FuncList{BiliBan.Filter_checkRecent, BiliBan.Filter_speed, BiliBan.Filter_checkModels},
		&BiliBan.ConfigMap{
			Filter_theSameCode_limit:    float32(0.45),
			Filter_speed_StartCheck:     2,
			Filter_speed_Limit:          float32(0.75),
			Filter_checkModels_limit:    float32(0.75),
			Filter_checkModels_models:   []string{},
			Filter_checkModels_expend:   []*BiliBan.RegVal{&BiliBan.RegVal{Compiled: reg1, Value: "#"}, &BiliBan.RegVal{Compiled: reg2, Value: ""}},
			Filter_checkRecent_limit:    0.9,
			Filter_checkRecent_length:   5,
			Filter_checkRecent_passtime: 60,
			Filter_keyword:              []string{"哔哩哔哩", "和你相遇"},
		})
	//创建热门房间
	PopularList, err := BiliBan.GetPopular(roomMax)
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
	var waitToChange []uint64
	go func() {
		for {
			select {
			case roomId := <-RoomCover:
				log.Println("进入等待列表")
				waitToChange = append(waitToChange, roomId)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-time.After(5 * time.Minute):
				if len(waitToChange) == 0 {
					continue
				}
				log.Printf("尝试补充房间，有%d个房间需要补充\n", len(waitToChange))
				oldRooms := BiliBan.UnitToMap(BiliBan.AllToUnit(&RoomRaw))
				PopularList, err = BiliBan.GetPopular(roomMax)
				if err != nil {
					log.Panic("丢失视野")
				}
				RoomRaw = PopularList.Get("data.#.roomid").Array()
				newRoomList := BiliBan.AllToUnit(&RoomRaw)
				for _, value := range *newRoomList {
					if BiliBan.InUint64Array(&waitToChange, value) {
						continue
					}
					if _, exits := (*oldRooms)[value]; !exits {
						enterRoom(&baseCtx, value, &msgIn, &RoomCover)
						pop := waitToChange[0]
						waitToChange = waitToChange[1:]
						log.Printf("自动切换房间，原房间%d，当前房间%d", pop, value)
						if len(waitToChange) == 0 {
							break
						}
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

package BiliBan

import (
	"context"
	"net"
	"sync"
	"time"
)

type LiveRoom struct {
	// 初始化属性
	RoomID     uint64          //房间ID，可短ID
	ReceiveMsg func(*MsgModel) //接受信息

	// 私有属性
	roomLongID uint64 //真ID

	conn     *net.TCPConn
	chBuffer chan *bufferInfo
	chMsg    chan *MsgModel
	ctx      context.Context
	cancel   context.CancelFunc
}

type MsgModel struct {
	UserID   uint64
	Level    uint64
	UserName string
	Content  string
	Ct       string //useForBan
	Time     int64
}

type bufferInfo struct {
	TypeID uint32
	Buffer []byte
}
type enterRoomModel struct {
	RoomID uint64 `json:"roomid"`
	UserID uint64 `json:"uid"`
}

type CheckCenter struct {
	BanRecords  []banRecord
	DanmuRecord sync.Map
	FuncConfig  map[string]interface{}
	msgConn     chan *MsgModel
	passFilter  []func(center *CheckCenter, model *MsgModel) bool
	banFilter   []func(center *CheckCenter, model *MsgModel) bool
	replaceMap  map[rune]rune
	//私有属性
	passTime  int //统计窗口大小
	minLength int //最小字符串长度
	danmuIn   int //弹幕入库数
}
type banRecord struct {
	BanTime time.Time
	*MsgModel
}
type FuncList []func(center *CheckCenter, model *MsgModel) bool
type Filter_speed_config struct {
	StartCheck int
	Limit      float32
}

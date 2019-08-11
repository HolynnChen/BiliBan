package BiliBan

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

const roomInfoURL string = "https://api.live.bilibili.com/room/v1/Room/room_init?id="
const dmServer string = "broadcastlv.chat.bilibili.com"
const dmPort int = 2243

func (room *LiveRoom) Start(ctx context.Context) {
	if err := room.init(); err != nil {
		log.Printf("房间%d获取信息失败", room.RoomID)
		log.Panic(err)
	}
	conn, err := room.createConnect()
	if err != nil {
		log.Printf("房间%d创建链接失败", room.RoomID)
		log.Panic(err)
	}
	log.Printf("房间%d初始化成功", room.RoomID)
	room.conn = <-conn
	room.chBuffer = make(chan *bufferInfo, 1000)
	room.chMsg = make(chan *MsgModel, 1000)
	room.ctx, room.cancel = context.WithCancel(ctx)
	defer room.cancel()
	go room.analysis(room.ctx)
	go room.distribute(room.ctx)
	room.enterRoom()
	go room.heartBeat(room.ctx)
	room.receive()
}
func (room *LiveRoom) init() error {
	resRoom, err := httpGetJsonWhitCheck(roomInfoURL + strconv.FormatUint(room.RoomID, 10))
	if err != nil {
		return err
	}
	if !Exits(resRoom, []string{"data.is_hidden", "data.is_locked", "data.room_id"}) {
		return errors.New(fmt.Sprintf("json结构不符合预期 %s", resRoom.Raw))
	}
	if resRoom.Get("data.is_hidden").Bool() || resRoom.Get("data.is_locked").Bool() {
		return errors.New("房间非法")
	}
	room.roomLongID = resRoom.Get("data.room_id").Uint()
	return nil
}
func (room *LiveRoom) createConnect() (<-chan *net.TCPConn, error) {
	result := make(chan *net.TCPConn)
	TcpConn, err := getConnect()
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(result)
		result <- TcpConn
	}()
	return result, nil
}
func getConnect() (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", dmServer+":"+strconv.Itoa(dmPort))
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, tcpAddr)
}
func (room *LiveRoom) sendData(action int, payload []byte) {
	b := bytes.NewBuffer([]byte{})
	_ = binary.Write(b, binary.BigEndian, int32(len(payload)+16))
	_ = binary.Write(b, binary.BigEndian, int16(16))
	_ = binary.Write(b, binary.BigEndian, int16(1))
	_ = binary.Write(b, binary.BigEndian, int32(action))
	_ = binary.Write(b, binary.BigEndian, int32(1))
	_ = binary.Write(b, binary.LittleEndian, payload)
	_, _ = room.conn.Write(b.Bytes())
}
func (room *LiveRoom) enterRoom() {
	enterInfo := &enterRoomModel{
		RoomID: room.roomLongID,
		UserID: 9999999999 + rand.Uint64(),
	}

	payload, err := json.Marshal(enterInfo)
	if err != nil {
		log.Panic(err)
	}
	room.sendData(7, payload)
}
func (room *LiveRoom) receive() {
	for {
		// 包头总长16个字节,包括 数据包长(4),magic(2),protocol_version(2),typeid(4),params(4)
		headBuffer := make([]byte, 16)
		_, err := io.ReadFull(room.conn, headBuffer)
		if err != nil {
			//log.Panicln(err)
			log.Println("出现故障，尝试自动恢复")
			conn, err := room.createConnect()
			if err != nil {
				log.Panic(err)
			}
			room.conn = <-conn
			room.enterRoom()

		}
		packetLength := binary.BigEndian.Uint32(headBuffer[:4])
		if packetLength < 16 || packetLength > 3072 {
			log.Println("***************协议失败***************")
			log.Println("数据包长度:", packetLength)
			conn, err := room.createConnect()
			if err != nil {
				log.Panic(err)
			}
			room.conn = <-conn
			room.enterRoom()
			continue
		}
		typeID := binary.BigEndian.Uint32(headBuffer[8:12]) // 读取typeid
		payLoadLength := packetLength - 16
		if payLoadLength == 0 {
			continue
		}
		payloadBuffer := make([]byte, payLoadLength)
		_, err = io.ReadFull(room.conn, payloadBuffer)
		if err != nil {
			log.Panicln(err)
		}
		room.chBuffer <- &bufferInfo{TypeID: typeID, Buffer: payloadBuffer}
	}
}
func (room *LiveRoom) distribute(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case value := <-room.chMsg:
			if room.ReceiveMsg != nil {
				if value.Level > 0 {
					continue
				}
				go room.ReceiveMsg(value)
			}
		}
	}
}
func (room *LiveRoom) analysis(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case buffer := <-room.chBuffer:
			switch buffer.TypeID {
			case 3:
				//viewer := binary.BigEndian.Uint32(buffer.Buffer)
			case 5:
				result := gjson.Parse(Buff2String(buffer.Buffer))
				if !result.Get("cmd").Exists() {
					continue
				}
				cmd := result.Get("cmd").String()
				switch cmd {
				case "PREPARING": //下播处理
					room.Preparing(room.RoomID)
				case "WELCOME":
				case "WELCOME_GUARD":
				case "DANMU_MSG":
					if result.Get("info.0.9").Uint() == 1 {
						continue
					}
					room.chMsg <- &MsgModel{
						UserID:   result.Get("info.2.0").Uint(),
						Level:    result.Get("info.4.0").Uint(),
						UserName: result.Get("info.2.1").String(),
						Content:  result.Get("info.1").String(),
						Ct:       result.Get("info.9.ct").String(),
						Time:     time.Now().Unix(),
					}
				case "SEND_GIFT":
				case "COMBO_END":
				case "GUARD_BUY":
				default:
					// log.Println(result.Data)
					// log.Println(string(buffer.Buffer))
				}
			default:
				break
			}
		}
	}
}
func (room *LiveRoom) heartBeat(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		room.sendData(2, []byte{})
		time.Sleep(30 * time.Second)
	}
}

package BiliBan

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"
)

func (center *CheckCenter) Init(passTime int, minLength int, passFilter FuncList, banFilter FuncList, FuncConfig *ConfigMap) chan *MsgModel {
	fmt.Println("封禁中心初始化")
	center.msgConn = make(chan *MsgModel, 10000)
	//赋值给私有变量
	center.passTime = passTime
	center.minLength = minLength
	center.danmuIn = 0
	center.config = FuncConfig
	center.passFilter = passFilter
	center.banFilter = banFilter
	center.replaceMap = map[rune]rune{}
	for _, kv := range strings.Split(RelaceString, ";") {
		temp := strings.Split(kv, ":")
		k, v := temp[0], []rune(temp[1])[0]
		for _, nv := range k {
			center.replaceMap[nv] = v
		}
	}
	return center.msgConn
}

func (center *CheckCenter) Start() {
	for i, max := 0, runtime.NumCPU(); i < max; i++ {
		go center.run()
	}
	center.autoClean()
}

//运行函数
func (center *CheckCenter) run() {
	for {
		select {
		case msg := <-center.msgConn:
			go center.check(msg)
		}
	}
}
func (center *CheckCenter) check(msg *MsgModel) {
	if utf8.RuneCountInString(msg.Content) < center.minLength {
		return
	}
	nowUserRecord, _ := center.DanmuRecord.LoadOrStore(msg.UserID, make([]*MsgModel, 0, 5))
	nowUserRecord = append(nowUserRecord.([]*MsgModel), msg)
	center.DanmuRecord.Store(msg.UserID, nowUserRecord)
	center.clean(msg.UserID)
	for _, function := range center.passFilter {
		if function(center, msg) {
			return
		}
	}
	for _, function := range center.banFilter {
		if function(center, msg) {
			center.ban(msg)
		}
	}

}
func (center *CheckCenter) clean(key uint64) {
	nowUserRecord, ok := center.DanmuRecord.Load(key)
	if !ok {
		return
	}
	newValue := make([]*MsgModel, 0)
	for _, record := range nowUserRecord.([]*MsgModel) {
		if time.Now().Unix()-record.Time <= int64(center.passTime) {
			newValue = append(newValue, record)
		}
	}
	if len(newValue) == 0 {
		center.DanmuRecord.Delete(key)
	} else {
		center.DanmuRecord.Store(key, newValue)
	}
}
func (center *CheckCenter) autoClean() {
	for {
		log.Println("触发自动清理")
		time.Sleep(time.Duration(center.passTime) * time.Second)
		go center.DanmuRecord.Range(func(k, _ interface{}) bool {
			center.clean(k.(uint64))
			return true
		})
	}
}

func (center *CheckCenter) ban(model *MsgModel) {
	if _, ok := center.preDel.Load(model.UserID); ok {
		return
	}
	center.preDel.Store(model.UserID, &struct{}{})
	fmt.Println("封禁")
	fmt.Println(model)
}

//判断函数区
func Levenshtein(s1 *string, s2 *string) float32 {
	r1, r2, result := utf8.RuneCountInString(*s1), utf8.RuneCountInString(*s2), ComputeDistance(*s1, *s2)
	return Min2(1-float32(result)/float32(r1), 1-float32(result)/float32(r2))
}
func ComputeDistance(a, b string) int {
	if len(a) == 0 {
		return utf8.RuneCountInString(b)
	}

	if len(b) == 0 {
		return utf8.RuneCountInString(a)
	}

	if a == b {
		return 0
	}
	s1 := []rune(a)
	s2 := []rune(b)

	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}
	lenS1 := len(s1)
	lenS2 := len(s2)
	x := make([]int, lenS1+1)
	for i := 0; i < len(x); i++ {
		x[i] = i
	}
	_ = x[lenS1]
	for i := 1; i <= lenS2; i++ {
		prev := i
		var current int
		for j := 1; j <= lenS1; j++ {
			if s2[i-1] == s1[j-1] {
				current = x[j-1] // match
			} else {
				current = min(min(x[j-1]+1, prev+1), x[j]+1)
			}
			x[j-1] = prev
			prev = current
		}
		x[lenS1] = prev
	}
	return x[lenS1]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

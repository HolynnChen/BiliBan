package BiliBan

import (
	"fmt"
	"runtime"
	"time"
	"unicode/utf8"
)

func (center *CheckCenter)Init(passTime int,minLength int, passFilter FuncList, banFilter FuncList,FuncConfig map[string]interface{}) chan *MsgModel {
	fmt.Println("封禁中心初始化")
	center.msgConn=make(chan *MsgModel,10000)
	//赋值给私有变量
	center.passTime=passTime
	center.minLength=minLength
	center.danmuIn=0
	center.FuncConfig=FuncConfig
	center.passFilter=passFilter
	center.banFilter=banFilter
	return center.msgConn
}

func (center *CheckCenter)Start(){
	for i,max:=0,runtime.NumCPU();i<max;i++{
		go center.run()
	}
}

//运行函数
func (center *CheckCenter)run(){
	for{
		select {
		case msg:=<-center.msgConn:
			fmt.Println(msg)
			go center.check(msg)
		}
	}
}
func (center *CheckCenter)check(msg *MsgModel){
	if utf8.RuneCountInString(msg.Content)>=center.minLength{
		nowUserRecord,_:=center.DanmuRecord.LoadOrStore(msg.UserID,make([]*MsgModel,0,5))
		nowUserRecord=append(nowUserRecord.([]*MsgModel),msg)
		center.DanmuRecord.Store(msg.UserID,nowUserRecord)
	}
	center.clean(msg.UserID)
	for _,function:=range center.passFilter{
		if function(center,msg){
			return
		}
	}
	for _,function:=range center.banFilter{
		if function(center,msg){
			center.ban(msg)
		}
	}

}
func (center *CheckCenter)clean(key uint64){
	nowUserRecord,ok:=center.DanmuRecord.Load(key)
	if !ok{
		return
	}
	newValue:=make([]*MsgModel,0)
	for _,record:=range nowUserRecord.([]*MsgModel) {
		if time.Now().Unix()-record.Time<=int64(center.passTime){
			newValue=append(newValue,record)
		}
	}
	if len(newValue)==0{
		center.DanmuRecord.Delete(key)
	}else{
		center.DanmuRecord.Store(key,newValue)
	}
}
func (center *CheckCenter)autoClean(){
	for{
		time.Sleep(30*time.Second)
	}
}

func (center *CheckCenter)ban(model *MsgModel){
	fmt.Println("封禁")
	fmt.Println(model)
}







//判断函数区
func Levenshtein(s1 *string,s2 *string)float32{
	rs1,rs2:=[]rune(*s1),[]rune(*s2)
	r1,r2:=len(rs1),len(rs2)
	dp:=make([]int,r1+1,r1+1)
	for i:=0;i<=r1;i++{
		dp[i]=i
	}
	for i:=1;i<=r2;i++{
		left,up:=i,i-1
		for j:=1;j<=r1;j++{
			cost:=0
			if rs1[j-1]==rs2[i-1]{cost=1}
			left=Min(dp[j]+1,left+1,up+cost)
			up=dp[j]
			dp[j]=left
		}
	}
	return Min2(1-float32(dp[r1])/float32(r1),1-float32(dp[r1])/float32(r2))
}
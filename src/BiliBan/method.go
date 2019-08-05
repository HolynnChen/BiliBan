package BiliBan

import "fmt"

func Filter_theSameCode(center *CheckCenter,model *MsgModel)bool{
	max,temp_map,strLen:=0,map[rune]int{},0
	for _,ch:=range model.Content{
		temp_map[ch]++
		strLen++
		max=Max(max,temp_map[ch])
	}
	if float32(max)/float32(strLen)>center.FuncConfig["BiliBan.Filter_theSameCode"].(float32){
		return true
	}
	return false
}
func Filter_speed(center *CheckCenter,model *MsgModel)bool{
	userMSG,has:=center.DanmuRecord.Load(model.UserID)
	if !has{
		return false
	}
	msgList:=userMSG.([]*MsgModel)
	fmt.Println(msgList)
	config:=center.FuncConfig["BiliBan.Filter_speed"].(Filter_speed_config)
	if len(msgList)>config.StartCheck{
		var allCompare float32=0.0
		for i,max:=1,len(msgList);i<max;i++{
			allCompare=(allCompare*float32(i-1)+Levenshtein(&msgList[i].Content,&msgList[i-1].Content))/float32(i)
			if i>2&&allCompare>config.Limit{
				return true
			}
		}
	}
	return false
}
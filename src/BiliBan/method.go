package BiliBan

import (
	"log"
	"strings"
	"time"
)

func Filter_theSameCode(center *CheckCenter, model *MsgModel) bool {
	max, tempMap, strLen := 0, map[rune]int{}, 0
	for _, ch := range model.Content {
		tempMap[ch]++
		strLen++
		max = Max(max, tempMap[ch])
	}
	if float32(max)/float32(strLen) > center.config.Filter_theSameCode_limit {
		return true
	}
	return false
}
func Filter_speed(center *CheckCenter, model *MsgModel) bool {
	userMSG, has := center.DanmuRecord.Load(model.UserID)
	if !has {
		return false
	}
	msgList := userMSG.([]*MsgModel)
	if len(msgList) > center.config.Filter_speed_StartCheck {
		var allCompare float32 = 0.0
		for i, max := 1, len(msgList); i < max; i++ {
			allCompare = (allCompare*float32(i-1) + Levenshtein(&msgList[i].Content, &msgList[i-1].Content)) / float32(i)
			if i > 2 && allCompare > center.config.Filter_speed_Limit {
				return true
			}
		}
	}
	return false
}

//字符串替换
func (center *CheckCenter) DanmuTransform(s *string) string {
	runeArray := []rune(*s)
	for index, value := range runeArray {
		if got, ok := center.replaceMap[value]; ok {
			runeArray[index] = got
			value = got
		}
		if value == 12288 {
			runeArray[index] = '#'
		} else if value > 65280 && value < 65375 {
			runeArray[index] = value - 65248
		} else if value >= 0x0030 && value <= 0x0039 ||
			value >= 0x2460 && value <= 0x249b ||
			value >= 0x3220 && value <= 0x3229 ||
			value >= 0x3248 && value <= 0x324f ||
			value >= 0x3251 && value <= 0x325f ||
			value >= 0x3280 && value <= 0x3289 ||
			value >= 0x32b1 && value <= 0x32bf ||
			value >= 0xff10 && value <= 0xff19 {
			runeArray[index] = '#'
		}
	}
	result := string(runeArray)
	for _, regVal := range center.config.Filter_checkModels_expend {
		result = regVal.Compiled.ReplaceAllString(result, regVal.Value)
	}
	return result
}

func Filter_checkModels(center *CheckCenter, model *MsgModel) bool {
	toCheck := center.DanmuTransform(&model.Content)
	for _, model := range center.config.Filter_checkModels_models {
		if Levenshtein(&model, &toCheck) > center.config.Filter_checkModels_limit {
			return true
		}
	}
	return false
}

func Filter_checkRecent(center *CheckCenter, model *MsgModel) bool {
	toCheck := center.DanmuTransform(&model.Content)
	num := len(center.BanRecords)
	for i, times := num-1, 0; i > num-1-center.config.Filter_checkRecent_length && i > -1; i-- {
		if time.Now().Unix()-center.BanRecords[i].BanTime > center.config.Filter_checkRecent_passtime {
			return false
		}
		waitCheck := center.DanmuTransform(&center.BanRecords[i].Content)
		if Levenshtein(&toCheck, &waitCheck) > center.config.Filter_checkRecent_limit {
			times++
		}
		if times > 1 {
			log.Println("窗口匹配成功")
			return true
		}
	}
	return false
}
func Filter_keyword(center *CheckCenter, model *MsgModel) bool {
	for i := 0; i < len(center.config.Filter_keyword); i++ {
		if strings.Contains(model.Content, center.config.Filter_keyword[i]) {
			return true
		}
	}
	return false
}

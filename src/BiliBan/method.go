package BiliBan

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
	return string(runeArray)
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

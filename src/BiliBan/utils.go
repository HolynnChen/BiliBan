package BiliBan

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strconv"
	"unsafe"
)

//封装http的Get方法
func httpGet(url string)([]byte,error){
	tr:=&http.Transport{TLSClientConfig:&tls.Config{InsecureSkipVerify:true}}
	client:=&http.Client{Transport:tr}
	resp,err:=client.Get(url)
	if err!=nil {return nil,err}
	defer resp.Body.Close()
	body,err:=ioutil.ReadAll(resp.Body)
	if err!=nil {return nil,err}
	return body,nil
}
func httpGetJson(url string)(gjson.Result,error){
	body,err:=httpGet(url)
	var newJson=gjson.Result{}
	if err!=nil{return newJson,err}
	newJson=gjson.Parse(*(*string)(unsafe.Pointer(&body)))
	return newJson,err
}
func httpGetJsonWhitCheck(url string)(gjson.Result,error){
	newJson,err:=httpGetJson(url)
	if err!=nil{return newJson,err}
	code:=newJson.Get("code")
	if !code.Exists(){
		return newJson,errors.New("状态码不存在")
	}else if val,err:=strconv.Atoi(code.String());err!=nil||val!=0{
		return newJson,errors.New(fmt.Sprintf("状态码非0，具体信息：%s",newJson.Raw))
	}
	return newJson,nil
}
func Exits(json gjson.Result,checks []string)bool{
	for _,key:=range checks{
		if !json.Get(key).Exists(){
			return false
		}
	}
	return true
}
func Buff2String(b []byte) string{
	return  *(*string)(unsafe.Pointer(&b))
}

func Min(base int,args...int)int{
	for _,v:=range args{
		if base>v{
			base=v
		}
	}
	return base
}
func Max(base int,args...int)int{
	for _,v:=range args{
		if base<v{
			base=v
		}
	}
	return base
}
func Min2(base float32,args...float32)float32{
	for _,v:=range args{
		if base>v{
			base=v
		}
	}
	return base
}
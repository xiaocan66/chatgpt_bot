package wxapi

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"

	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

type WxBot struct {
	// WxReceiveFunc 消息交付
	WxReceiveFunc func(msg WxReceiveCommonMsg)
	APPID         string
	APPSECRET     string
	Port          string
	accessToken   string
	lock          sync.RWMutex
}
type token struct {
	AccessToken string `json:"access_token"`

	ExpiresIn int `json:"expires_in"`
}

func (v *WxBot) Login() {

	// 定时获取accessToken
	go func() {
		for {
			func() {
				v.lock.Lock()
				defer v.lock.Unlock()
				v.accessToken = v.wxGetAccessToken()

			}()

			time.Sleep(time.Minute * 5)

		}

	}()

	go func() {
		engine := gin.Default()

		engine.GET("/", v.handleWxLogin)
		engine.POST("/", v.handleWxPostRecv)
		if err := engine.Run(v.Port); err != nil {
			log.Fatal(err)
		}

	}()

}

// HandleWxLogin 首次接入，成为开发者
func (v *WxBot) handleWxLogin(c *gin.Context) {
	fmt.Printf("==>HandleWxLogin\n")
	echoStr := c.DefaultQuery("echostr", "")
	if echoStr != "" {
		fmt.Printf("==>echostr:%s\n", echoStr)
		c.String(200, "%s", echoStr)
		return
	}

}

type WxReceiveCommonMsg struct {
	ToUserName   string // 接收者ID
	FromUserName string // 发送者ID
	Content      string
	CreateTime   int64
	MsgType      string
	MsgId        int64
	PicUrl       string
	MediaId      string
	Event        string
	EventId      string
	Format       string
	Recognition  string
	ThumbMediaID string
}

type wxCustomText struct {
	Content string `json:"content"`
}
type WxCustomTextMsg struct {
	ToUser  string       `json:"touser"`
	MsgType string       `json:"msgtype"`
	Text    wxCustomText `json:"text"`
}

func (msg *WxCustomTextMsg) toJson() []byte {
	body, err := json.Marshal(msg)
	if err != nil {
		panic(err)

	}
	return body
}

func (v *WxBot) wxMakeSign(token, timestamp, nonce string) string {
	strs := []string{token, timestamp, nonce}
	sort.Strings(strs)
	sha := sha1.New()
	_, err := io.WriteString(sha, strings.Join(strs, ""))
	if err != nil {
		log.Println(err)
		return ""
	}
	return fmt.Sprintf("%x", sha.Sum(nil))
}
func (v *WxBot) GetAccessToken() string {
	v.lock.RLock()
	defer v.lock.RUnlock()
	return v.accessToken
}

// WxGetAccessToken 获取微信accessToken
func (v *WxBot) wxGetAccessToken() string {

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%v&secret=%v", v.APPID, v.APPSECRET)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取微信token失败", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("微信token读取失败", err)
		return ""
	}

	token := token{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		fmt.Println("微信token解析json失败", err)
		return ""
	}

	return token.AccessToken
}

// 获取关注列表
func (v *WxBot) wxGetUserList(accessToken string) []gjson.Result {
	url := "https://api.weixin.qq.com/cgi-bin/user/get?access_token=" + accessToken + "&next_openid="
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取关注列表失败", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return nil
	}
	flist := gjson.Get(string(body), "data.openid").Array()
	return flist

}

func (v *WxBot) wxPostTemplate(accessToken string, reqdata string, fxurl string, templateId string, openid string) {
	url := "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=" + accessToken
	reqbody := "{\"touser\":\"" + openid + "\", \"template_id\":\"" + templateId + "\", \"url\":\"" + fxurl + "\", \"data\": " + reqdata + "}"
	fmt.Printf("WxPostTemplate:%#v\n", reqbody)
	resp, err := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader(string(reqbody)))
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))
}

// 客服回复接口
func (v *WxBot) WxPostCustomTextMsg(accessToken string, touser string, content string) {
	if touser == "" || content == "" || accessToken == "" {
		return
	}
	url := "https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=" + accessToken

	req := &WxCustomTextMsg{ToUser: touser, MsgType: "text", Text: wxCustomText{Content: content}}
	jsonStr := req.toJson()
	//fmt.Printf("WxPostCustomTextMsg:%#v\n", jsonStr)
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))
}

// ReceiveCommonMsg
func (v *WxBot) receiveCommonMsg(msgData []byte) (WxReceiveCommonMsg, error) {

	fmt.Printf("received weixin msgData:\n%s\n", msgData)
	msg := WxReceiveCommonMsg{}
	err := xml.Unmarshal(msgData, &msg)
	if v.WxReceiveFunc == nil {
		return msg, err
	}

	v.WxReceiveFunc(msg)
	return msg, err
}

// HandleWxPostRecv 处理微信公众号前端发起的消息事件
func (v *WxBot) handleWxPostRecv(c *gin.Context) {
	fmt.Printf("==>HandleWxPostRecv Enter\n")
	data, err := c.GetRawData()
	if err != nil {
		log.Fatalln(err)
	}
	v.receiveCommonMsg(data)
}

// WxCreateMenu 创建菜单
func (v *WxBot) wxCreateMenu(accessToken, menustr string) (string, error) {

	url := "https://api.weixin.qq.com/cgi-bin/menu/create?access_token=" + accessToken
	fmt.Printf("WxCreateMenu:%s\n", menustr)
	resp, err := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader(menustr))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	fmt.Println(string(body))
	return string(body), nil

}

// WxDelMenu 删除菜单
func (v *WxBot) wxDelMenu(accessToken string) (string, error) {
	url := "https://api.weixin.qq.com/cgi-bin/menu/delete?access_token=" + accessToken
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("删除菜单失败", err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return "", err
	}

	fmt.Println(string(body))
	return string(body), nil

}

// WxGetUserInfo 根据用户openid获取基本信息
func (v *WxBot) wxGetUserInfo(accessToken, openid string) (string, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN", accessToken, openid)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取信息失败", err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return "", err
	}

	fmt.Println(string(body))
	return string(body), nil

}

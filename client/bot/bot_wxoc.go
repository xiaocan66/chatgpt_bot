package bot

import (
	"chatgpt-bot/cfg"
	"chatgpt-bot/engine"
	"chatgpt-bot/wxapi"
	"fmt"
	"log"
	"sync"
	"time"
)

type WxOcBot struct {
	engine engine.Engine
	wxBot  wxapi.WxBot
}

func NewWxOcBot() *WxOcBot {
	return &WxOcBot{}
}

func (w *WxOcBot) Init(cfg *cfg.Config) error {
	w.Register(cfg.WxOC.AppID, cfg.WxOC.AppSecret, cfg.WxOC.Port)
	w.engine = engine.GetEngine(cfg.EngineConfig.EngineType)
	err := w.engine.Init(cfg)
	if err != nil {
		return err
	}

	return nil
}

func (w *WxOcBot) Register(appid, appSecret, port string) {
	w.wxBot.WxReceiveFunc = w.handleWxMessage
	if appid == "" {
		panic("appid error")
	}
	if appSecret == "" {
		panic("appsecret  err")
	}
	if port == "" {
		panic("client port error")
	}
	w.wxBot.APPID = appid
	w.wxBot.APPSECRET = appSecret
	w.wxBot.Port = port

}

var msgMap sync.Map
var limit map[string]time.Time = make(map[string]time.Time)
var lock sync.Mutex

func (w *WxOcBot) handleWxMessage(msg wxapi.WxReceiveCommonMsg) {
	fmt.Println("weixin msg received")
	fmt.Printf("%#v\n", msg)

	touser := msg.FromUserName
	// 订阅消息
	if msg.Event == "subscribe" {
		welcome := "你好，我是ChatGPT，是一款由OpenAI开发的人工智能语言模型。我的任务是为用户提供各种语言相关的服务和答案，我可以回答关于文化、科学、历史、社会和技术等各种领域的问题，同时我也能够进行自然语言处理、文本生成和语言翻译等任务。\n\n我的工作原理是基于大规模机器学习技术，我从海量的语料库中学习语言知识，通过深度学习算法不断提高自身的语言理解和生成能力。我可以使用各种语言，如英语、汉语、西班牙语、法语等，来与用户进行交流和沟通。\n\n我被广泛应用于各种领域，例如教育、医疗、金融、商业等等。无论你是需要帮助解决问题，还是需要进行语言交流和翻译，我都可以为你提供支持。"

		w.wxBot.WxPostCustomTextMsg(w.wxBot.GetAccessToken(), touser, welcome)
		welcome = "如果您想了解某个特定领域的知识，您可以直接向我提问，我会尽可能给出准确和详细的答案。如果您需要某个领域的专业指导或者深入讨论，您也可以提出来，我会尽力为您提供支持和协助。\n\n如果您需要进行文本生成，例如新闻报道、文学作品、诗歌、电影剧本等等，您可以向我提供相关的输入内容和格式要求，我会尽力生成符合您要求的文本。当然，由于我的生成能力是基于机器学习技术，所以有时候可能会存在一些错误或者不太符合您期望的地方，但我会尽可能优化和改进，提供更加满意的结果。\n\n如果您需要进行语言翻译，我也可以为您提供支持。只需要提供源语言和目标语言，我就可以帮助您进行翻译。无论是常用语言，如英语、汉语、西班牙语、法语等等，还是一些较为罕见的语言，我都会尽可能提供准确和流畅的翻译服务。\n总之，无论您有什么需要，只要是涉及到语言和文本方面的问题和需求，我都会尽可能为您提供支持和解答。"
		w.wxBot.WxPostCustomTextMsg(w.wxBot.GetAccessToken(), touser, welcome)

		return
	}

	content := msg.Content

	if _, ok := msgMap.Load(touser); ok {
		w.wxBot.WxPostCustomTextMsg(w.wxBot.GetAccessToken(), touser, "😅你已经发送了一条信息，ChatGPT正在努力生成结果中,请耐心等待......")
		return
	}
	lock.Lock()
	defer lock.Unlock()
	l, ok := limit[touser]
	if ok && time.Since(l) < 2*time.Minute {

		w.wxBot.WxPostCustomTextMsg(w.wxBot.GetAccessToken(), touser, "为保证服务稳定运行,限制每位用户每次生成内容的间隔为2分钟! 下次可生成时间:"+l.Add(time.Minute*2).Format("2006-01-02 15:04:05"))
		return
	}
	limit[touser] = time.Now()
	w.wxBot.WxPostCustomTextMsg(w.wxBot.GetAccessToken(), touser, "😅ChatGPT正在努力生成结果中,请耐心等待......")
	go func() {
		msgMap.Store(touser, msg.Content)
		defer msgMap.Delete(touser)

		resp, err := w.engine.Chat(content)
		// 重新生成token  避免过期

		if err != nil {
			log.Println(err)
		}

		if resp != "" {
			//每条消息最大八百字
			textRune := []rune(resp)
			var left = 0
			var right = 800

			for {
				if right < len(textRune) {
					w.wxBot.WxPostCustomTextMsg(w.wxBot.GetAccessToken(), touser, string(textRune[left:right]))
					left = right
					right += 800
				}
				if right >= len(textRune) {
					w.wxBot.WxPostCustomTextMsg(w.wxBot.GetAccessToken(), touser, string(textRune[left:]))
					break
				}
			}

		} else {
			w.wxBot.WxPostCustomTextMsg(w.wxBot.GetAccessToken(), touser, "chatGPT服务异常,请联系微信:lizican123 提交bug..")
		}
	}()
}

func (w *WxOcBot) Run() {
	w.wxBot.Login()
}

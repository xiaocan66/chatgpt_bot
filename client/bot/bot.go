package bot

import "chatgpt-bot/cfg"

var (
	Telegram = "telegram"
	Wechat   = "wechat"
	WxOc     = "wxoc"
)

type Bot interface {
	Init(*cfg.Config) error
	Run()
}

func GetBot(botType string) Bot {
	switch botType {
	case Wechat:
		return NewWechatBot()
	case Telegram:
		return NewTelegramBot()
	case WxOc:
		return NewWxOcBot()
	default:
		return NewTelegramBot()
	}
}

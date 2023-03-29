# chatgpt_bot
chatgpt 接入微信公众号 ,Telegram,微信客户端 

```
chatgpt_bot
├── chat_gpt
│   ├── Dockerfile
│   ├── config.yaml ## 需要修改的配置文件
│   ├── main.py
│   ├── requirements.txt
│   └── session
│       └── session.py
├── client
│   ├── Dockerfile
│   ├── app
│   │   └── app.go
│   ├── bot
│   │   ├── bot.go
│   │   ├── bot_telegram.go
│   │   ├── bot_wechat.go
│   │   └── bot_wxoc.go
│   ├── build.sh
│   ├── cfg
│   │   └── config.go
│   ├── config.yaml  ##需要修改的配置文件
│   ├── config.yaml.example
│   ├── constant
│   │   ├── command.go
│   │   ├── constant.go
│   │   └── error.go
│   ├── engine
│   │   ├── engine.go
│   │   ├── engine_bing.go
│   │   └── engine_chatgpt.go
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── middleware
│   │   └── ratelimiter.go
│   ├── model
│   │   └── chat_task.go
│   ├── utils
│   │   ├── json_util.go
│   │   ├── string_util.go
│   │   └── string_util_test.go
│   └── wxapi
│       └── wxapi.go
├── docker-compose.yml
└── run.sh
```
## 如何部署
### 首先需要安装Docker环境
```shell
curl -fsSL https://get.docker.com | bash -s docker 
```
### 修改配置文件
配置文件①
```yaml
#path:chatgpt_bot/chat_gpt/config.yaml
chatgpt:
  host: "localhost"
  port: "5000"
  debug: True
  tokens:
      - 账号:密码
      - 账号:密码
```
配置文件②
```yaml
##path: chatgpt_bot/client/config.yaml

## 微信公众号作为入口  需要完成微信认证  个人实名的不行
bot:
  type: wxoc
  wxoc:
    appid: ''
    app_secret: ''
    port: ':8070' # 必须带':'
engine:
  host: engine
  port: 5000
  type: chatgpt


## 微信作为入口
# bot:
#   type: wechat 


# engine:
#   host: engine
#   port: 5000
#   type: chatgpt



#telegram 作为入口
# bot:
#   type: telegram
#   telegram:
#      token: your bot token
#      channelName:
#      groupName: @your group name
# engine:
#   host: engine
#   port: 5000
#   type: chatgpt

```
## 开始部署
```shell
sh run.sh
```


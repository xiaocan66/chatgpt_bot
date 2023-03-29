import logging

import yaml
from OpenAIAuth import Error as OpenAIError
from flask import Flask, request

from session.session import Session


def load_conf():
    try:
        conf_path = "./config.yaml"
        with open(conf_path, "r", encoding="utf-8") as f:
            return yaml.load(f, yaml.FullLoader)

    except Exception as e:
        logging.error(f'load config error {e}')
        exit(1)


chatgpt = None
app = Flask(__name__)


def main():
    conf = load_conf()
    print(conf)
    global chatgpt

    chatgpt = Session(config=conf)
    port = conf['chatgpt']['port']
    debug = conf['chatgpt'].get('debug', False)
    app.run(host='0.0.0.0', port=port, debug=debug)


@app.route("/chat")
def chat():
    sentence = request.args.get("sentence")
    logging.info(f"[Engine] chat gpt engine get request: {sentence}")
    try:
        res = chatgpt.chat_with_chatgpt(sentence)
        logging.info(f"[Engine] chat gpt engine get response: {res}")
        return {"message": res}
    except OpenAIError as e:
        logging.error(
            "[Engine] chat gpt engine get open api error: status: {}, details: {}".format(e.status_code, e.details))
        return {"detail": e.details, "code": e.status_code}
    except Exception as e:
        logging.error(f"[Engine] chat gpt engine get error: {str(e)}")
        return {"detail": str(e)}
@app.route("/reset")
def reset():
    chatgpt.clear()

if __name__ == "__main__":
    # server = pywsgi.WSGIServer(('0.0.0.0', 5000), app, handler_class=WebSocketHandler)
    # server.serve_forever()
    main()


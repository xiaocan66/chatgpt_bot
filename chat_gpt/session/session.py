import logging
import multiprocessing
import time

from revChatGPT.V1 import Chatbot as ChatGPTBot


class Session:
    def __init__(self, config):
        tokens = list(map(Session.map_token_dict, config["chatgpt"]["tokens"]))
        self.queue = multiprocessing.Queue(maxsize=len(tokens))

        self.init_chat_gpt_bot_with_credential(tokens)
        self.verbose = config["chatgpt"].get('debug', False)

    @staticmethod
    def map_token_dict(tokens):
        credentials = tokens.split(":")
        length = len(credentials)
        if length != 2 and length != 3:
            raise Exception("token format error")
        token = {"email": credentials[0], "password": credentials[1]}
        return token

    def chat_with_chatgpt(self, sentence: str):

        try:
            for i in range(0, 2):
                chat = self.queue.get(timeout=1000 * 60 * 10)
                time.sleep(3)

                len_text = 0
                msg = ""
                for data in chat.ask(sentence):
                    message = data['message'][len_text:]
                    len_text += len(message)
                    msg += message
                    logging.info(msg)
                return msg
        except Exception as e:
            logging.error(e)
            chat.reset_chat()
            chat.clear_conversations()
            return "服务器错误,请联系管理员微信:lizican123 提交bug!"
        finally:
            self.queue.put(chat)
        return "服务器错误,请联系管理员微信:lizican123 提交bug!"

    def init_chat_gpt_bot_with_credential(self, tokens):
        for token in tokens:
            try:
                gpt = ChatGPTBot(
                    config={
                        "email": token["email"],
                        "password": token["password"],


                    },
                    lazy_loading=True

                )
            except:
                continue

            self.queue.put(gpt)

    def clear(self):
        for i in range(0, 2):
            chat = self.queue.get(timeout=1000 * 60 * 10)
            chat.clear_conversations()
            self.queue.put(chat)

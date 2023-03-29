FROM golang:alpine

WORKDIR /app

COPY client/. .
RUN apk update && apk add tzdata 
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime 
RUN echo "Asia/Shanghai" > /etc/timezone

RUN go mod tidy 

RUN go build -ldflags '-w -s' -o main .

CMD ["./main"]

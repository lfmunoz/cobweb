# 839MB
#FROM golang:1.15.5  
FROM golang:1.15.5-alpine

ENV GO111MODULE=on

WORKDIR /root

ADD web /go/bin/web
ADD conf.json /go/bin/conf.json
ADD main /go/bin/main
RUN chmod +x /go/bin/main

EXPOSE 8080

CMD ["main"]
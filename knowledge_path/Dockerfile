FROM alpine

COPY main.go /root/go/src/knowledge_path/main.go
RUN apk add go git g++

RUN go get  github.com/tidwall/gjson && go get github.com/wilcosheh/tfidf 

RUN go build knowledge_path

EXPOSE 80/tcp

ENTRYPOINT /knowledge_path


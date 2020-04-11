FROM alpine

COPY main.go /root/go/src/knowledge_graph/main.go
RUN apk add go git

RUN go get  github.com/tidwall/gjson && go build knowledge_graph

RUN apk del go git

EXPOSE 80/tcp

ENTRYPOINT /knowledge_graph


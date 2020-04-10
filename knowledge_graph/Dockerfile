FROM alpine

COPY main.go /root/go/src/knowledge_graph/main.go
RUN apk add go

RUN go build knowledge_graph

RUN apk del go

EXPOSE 80/tcp

ENTRYPOINT /knowledge_graph


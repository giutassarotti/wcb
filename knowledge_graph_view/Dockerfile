FROM alpine

COPY main.go /root/go/src/knowledge_graph_view/main.go
RUN apk add go

RUN go build knowledge_graph_view

RUN apk del go

EXPOSE 80/tcp

ENTRYPOINT /knowledge_graph_view


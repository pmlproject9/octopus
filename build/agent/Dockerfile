FROM golang:1.19 as builder

ENV CGO_ENABLED=0 GOOS=linux
ENV GOPROXY=https://goproxy.cn,direct
ENV GOCACHE=/gocache

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY build/ build/

RUN go mod download

RUN id=gobuild,target=/gocache \
    go build -v -o /agent ./cmd/octopus-agent/main.go

FROM alpine:3.13.4
RUN apk add ipset && apk add --no-cache iptables  && apk add ip6tables
WORKDIR /
COPY --from=builder /agent .
COPY --from=builder /workspace/build/agent/iptables-wrapper-installer.sh /

RUN /iptables-wrapper-installer.sh --no-sanity-check

ENTRYPOINT ["/agent"]



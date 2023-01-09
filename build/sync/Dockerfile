FROM golang:1.19 as builder

ENV CGO_ENABLED=0 GOOS=linux
ENV GOPROXY=https://goproxy.cn,direct
ENV GOCACHE=/gocache

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
COPY cmd/ cmd/
COPY pkg/ pkg/

RUN go mod download

RUN --mount=type=cache,id=gobuild,target=/gocache \
    go build -v -o /sync ./cmd/service-sync/main.go

FROM alpine:3.13.4
WORKDIR /
COPY --from=builder /sync .
ENTRYPOINT ["/sync"]
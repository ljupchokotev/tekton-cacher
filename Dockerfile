FROM golang:1.17.5

WORKDIR /go/src/tekton-cacher

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ./

RUN go build -o /usr/bin/store ./cmd/store
RUN go build -o /usr/bin/restore ./cmd/restore

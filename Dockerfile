FROM golang:1.15-alpine AS builder

RUN apk add --no-cache git

WORKDIR $GOPATH/src/github.com/Vilsol/yeet/

ENV GO111MODULE=on

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /yeet main.go

FROM scratch

COPY --from=builder /yeet /yeet

ENTRYPOINT ["/yeet"]
FROM golang:1.24-alpine AS deps

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

FROM golang:1.24-alpine AS builder

WORKDIR /usr/src/app
COPY --from=deps /go/pkg /go/pkg

ENV GODEBUG=netdns=go

ENV PORT=8080
ENV GIN_MODE=release

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -v -o /usr/local/bin/app ./main.go

CMD ["app"]

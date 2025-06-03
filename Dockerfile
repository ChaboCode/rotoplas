FROM golang:1.24-alpine

WORKDIR /usr/src/app

ENV PORT=8080
ENV GIN_MODE=release

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -v -o /usr/local/bin/app ./main.go

CMD ["app"]

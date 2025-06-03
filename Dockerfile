FROM golang:latest

WORKDIR /usr/src/app

ENV PORT=80

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/local/bin/app ./main.go

CMD ["app"]

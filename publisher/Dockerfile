FROM golang:1.25.3-alpine

WORKDIR /app

RUN apk add --no-cache git bash

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o publisher .

EXPOSE 4000

CMD ["./publisher"]

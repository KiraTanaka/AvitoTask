FROM golang:1.23.1-alpine AS build

WORKDIR /avitoTask

RUN apk --no-cache add bash make gcc gettext musl-dev

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN ls

RUN go build ./cmd/main.go

EXPOSE 8080

CMD ["./main"]


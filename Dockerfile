FROM golang:1.23.1-bullseye AS build

WORKDIR /avitoTask

#RUN apk --no-cache add bash make gcc gettext musl-dev

COPY . .

COPY go.mod go.sum ./

RUN go mod download

RUN ls

RUN go build ./cmd/main.go

EXPOSE 8080

RUN chmod +x ./entypoint.sh

CMD ["./main"]


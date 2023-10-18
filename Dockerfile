FROM golang:latest AS builder

WORKDIR /backend
COPY ./backend .

RUN go install github.com/cosmtrek/air@latest
RUN go mod download

CMD ["air", "-c", ".air.toml"]

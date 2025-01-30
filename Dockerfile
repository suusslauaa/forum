FROM golang:1.23.4-alpine
RUN apk add --no-cache build-base gcc
WORKDIR /forum
COPY . .
RUN apk add --no-cache sqlite
RUN go build -o forum cmd/main.go

CMD ["./forum"]

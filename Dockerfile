FROM golang:1.21-alpine
WORKDIR /app
COPY backend ./backend
WORKDIR /app/backend
RUN go mod init github.com/Marketen/POC-beaconchainAPI/backend || true
RUN go mod tidy
RUN go build -o app ./cmd/server/main.go
CMD ["./app"]

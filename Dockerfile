FROM docker.io/library/golang:1.23 as builder
WORKDIR /app/
COPY go.mod go.sum /app/
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v

FROM alpine
COPY --from=builder /app/deepseek-telegram-bot /app/deepseek-telegram-bot

ENTRYPOINT ["/app/deepseek-telegram-bot"]
ENV DS_API_KEY= BOT_TOKEN= DS_INITIAL_PROMPT= DS_TEMPERATURE= DS_MAX_REPLY_TOKENS= DS_HISTORY_SIZE= ALLOWED_USERIDS= ADMIN_USERIDS= ALLOWED_GROUPIDS=

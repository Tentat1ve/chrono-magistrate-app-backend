FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/dignitarie_lists

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
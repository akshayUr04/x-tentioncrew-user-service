FROM golang:1.21-alpine3.18 AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache git \
    && go mod download \
    && go build -o ./out/exe cmd/main.go

FROM alpine:3.18
COPY --from=builder /app/out/exe /app/
WORKDIR /app
EXPOSE 50052
CMD ["./exe"]

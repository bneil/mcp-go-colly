# Build stage ------------------------------------------------------------
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache make git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

# Final stage ------------------------------------------------------------
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/bin/mcp-go-colly /app/bin/
COPY smithery.yaml /app/

USER appuser

ENTRYPOINT ["/app/bin/mcp-go-colly"] 
FROM golang:1.26.2-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./internal/migrate

FROM alpine:3.22

RUN addgroup -S app && adduser -S -G app app \
	&& apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /out/server ./server
COPY --from=builder /out/migrate ./migrate

USER app
EXPOSE 8080

ENTRYPOINT ["/app/server"]

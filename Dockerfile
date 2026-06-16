FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.mod ./
COPY . .

RUN go build -ldflags="-s -w" -o yebobank ./cmd/server

FROM alpine:3.19

RUN addgroup -S yebo && adduser -S yebo -G yebo

WORKDIR /app

COPY --from=builder /build/yebobank .
COPY --from=builder /build/web ./web
COPY --from=builder /build/docs/database/migrations ./migrations

RUN chown -R yebo:yebo /app

USER yebo

EXPOSE 8080

ENTRYPOINT ["./yebobank"]

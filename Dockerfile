# =====================================
# Etapa 1: Builder
# =====================================

FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o sdata-ms .

# =====================================
# Etapa 2: Producción
# =====================================

FROM alpine:3.22

RUN apk --no-cache add ca-certificates && \
    addgroup -g 10001 -S appgroup && \
    adduser -S -D -H -u 10001 -G appgroup appuser

WORKDIR /app

COPY --from=builder /app/sdata-ms .

RUN chown -R 10001:10001 /app

USER 10001:10001

EXPOSE 50051

CMD ["./sdata-ms"]

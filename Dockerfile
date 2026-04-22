# ========== Build stage ==========
FROM golang:1.25-alpine AS build
RUN apk add --no-cache build-base libwebp-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=1
RUN go build -ldflags='-s -w' -o /out/vouchers ./cmd/vouchers

# ========== Runtime stage ==========
FROM alpine:3.20
RUN apk add --no-cache \
      ca-certificates \
      tzdata \
      libwebp \
      wget && \
    adduser -D -u 10001 -s /sbin/nologin vouchers && \
    mkdir -p /srv/vouchers && \
    chown -R vouchers:vouchers /srv/vouchers
COPY --from=build /out/vouchers /usr/local/bin/vouchers
USER vouchers
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/vouchers"]

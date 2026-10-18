# syntax=docker/dockerfile:1

# Stage 1: Build from source
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH
ARG VERSION=dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /build/vivechak ./cmd/vivechak && \
    ln -s /build/vivechak /build/vck

# Stage 2: Minimal runtime container
FROM alpine:3.20

COPY --from=builder /build/vivechak /usr/local/bin/vivechak
COPY --from=builder /build/vck /usr/local/bin/vck

RUN chmod +x /usr/local/bin/vivechak /usr/local/bin/vck && \
    adduser -D -u 10001 -s /bin/sh vivechak

WORKDIR /workspace
USER vivechak:vivechak

# Default entrypoint runs MCP server over stdio with zero banner noise
ENTRYPOINT ["/usr/local/bin/vivechak"]
CMD ["serve"]

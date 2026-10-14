# syntax=docker/dockerfile:1
FROM alpine:3.20

ARG TARGETARCH

# Copy cross-compiled binaries matching the specific target platform
COPY docker-dist/${TARGETARCH}/vivechak /usr/local/bin/vivechak
COPY docker-dist/${TARGETARCH}/vck /usr/local/bin/vck

RUN chmod +x /usr/local/bin/vivechak /usr/local/bin/vck && \
    adduser -D -u 10001 -s /bin/sh vivechak

WORKDIR /workspace
USER vivechak:vivechak

# Default entrypoint runs MCP server over stdio with zero banner noise
ENTRYPOINT ["/usr/local/bin/vivechak"]
CMD ["serve"]

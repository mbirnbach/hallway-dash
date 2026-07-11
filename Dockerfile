### Build stage ###
# Built on the native (host) platform and cross-compiled to the target
# platform via GOOS/GOARCH — much faster than QEMU-emulating the whole
# Go toolchain per architecture in a multi-arch buildx build.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/hallway-dash .

### Runtime stage ###
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S hallway && adduser -S hallway -G hallway
WORKDIR /app

COPY --from=build /out/hallway-dash ./hallway-dash
COPY static ./static

# Mounted at runtime via the compose volume / Unraid template path
# mapping; created here so the container has a valid default even if
# the mount is omitted.
RUN mkdir -p /app/backgrounds && chown -R hallway:hallway /app

ENV STATIC_DIR=/app/static
ENV BACKGROUNDS_DIR=/app/backgrounds
ENV PORT=8080

USER hallway
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/api/config || exit 1

ENTRYPOINT ["./hallway-dash"]

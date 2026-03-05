FROM --platform=$BUILDPLATFORM golang:alpine AS build-0

# Set up working directory
WORKDIR /app

# Copy go.mod and go.sum separately so we only invalidate the downloading layers if we need to
COPY go.mod go.sum ./

# Fetch dependencies and build the binary
ENV CGO_ENABLED=0

RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go mod download -x

# Copy the rest of the project to ensure code changes doesnt trigger re-download of all deps
COPY . .

FROM --platform=$BUILDPLATFORM build-0 AS build-sleep

ARG TARGETOS
ARG TARGETARCH
ARG GIT_TAG

RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -o sleep_${TARGETOS}_${TARGETARCH} -ldflags="-w -s -X main.version=${GIT_TAG}" ./cmd/sleep


FROM --platform=$BUILDPLATFORM build-0 AS build-sftp

ARG TARGETOS
ARG TARGETARCH
ARG GIT_TAG

RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -o podsftp_${TARGETOS}_${TARGETARCH} -ldflags="-w -s -X main.version=${GIT_TAG}" ./cmd/podsftp


# Runner 
FROM scratch

# Set up the working directory
WORKDIR /

ARG TARGETOS
ARG TARGETARCH

# Copy the binary from the builder stage
COPY --from=build-sleep --chmod=777 /app/sleep_${TARGETOS}_${TARGETARCH} /usr/bin/sleep
COPY --from=build-sftp --chmod=777 /app/podsftp_${TARGETOS}_${TARGETARCH} /usr/bin/podsftp

ENTRYPOINT ["/usr/bin/podsftp"]


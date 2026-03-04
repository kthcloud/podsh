FROM alpine:latest AS runner-setup

RUN addgroup -S syncdb \
 && adduser  -S syncdb \
             -G syncdb \
             -h /opt/syncdb \
             -s /sbin/nologin \
             -D \
 && mkdir -p /opt/syncdb \
 && chown syncdb:syncdb /opt/syncdb

FROM --platform=$BUILDPLATFORM golang:alpine AS build-0

# Set up working directory
WORKDIR /app

# Copy go.mod and go.sum separately so we only invalidate the downloading layers if we need to
COPY go.mod go.sum ./

# Fetch dependencies and build the binary
ENV CGO_ENABLED=0

RUN go mod download -x

# Copy the rest of the project to ensure code changes doesnt trigger re-download of all deps
COPY . .

FROM --platform=$BUILDPLATFORM build-0 AS build-syncdb

ARG TARGETOS
ARG TARGETARCH
ARG GIT_TAG

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -o syncdb_${TARGETOS}_${TARGETARCH} -ldflags="-w -s -X main.version=${GIT_TAG}" ./cmd/syncdb

# Runner 
FROM scratch

COPY --from=runner-setup /etc/passwd /etc/group /etc/passwd
COPY --from=runner-setup /etc/group /etc/group

ENV HOME=/opt/syncdb

# Set up the working directory
WORKDIR /opt/syncdb

ARG TARGETOS
ARG TARGETARCH

# Copy the binary from the builder stage
COPY --from=build-syncdb --chmod=777 /app/syncdb_${TARGETOS}_${TARGETARCH} /usr/bin/syncdb

VOLUME /etc/syncdb

ENTRYPOINT ["/usr/bin/syncdb"]


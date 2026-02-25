FROM alpine:latest AS runner-setup

RUN addgroup -S podsh \
 && adduser  -S podsh \
             -G podsh \
             -h /opt/podsh \
             -s /sbin/nologin \
             -D \
 && mkdir -p /opt/podsh \
 && chown podsh:podsh /opt/podsh

FROM --platform=$BUILDPLATFORM golang:alpine AS build-0

RUN apk add --no-cache libcap

# Set up working directory
WORKDIR /app

# Copy go.mod and go.sum separately so we only invalidate the downloading layers if we need to
COPY go.mod go.sum ./

# Fetch dependencies and build the binary
ENV CGO_ENABLED=0

RUN go mod download -x

# Copy the rest of the project to ensure code changes doesnt trigger re-download of all deps
COPY . .

FROM --platform=$BUILDPLATFORM build-0 AS build-podsh

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -o podsh_${TARGETOS}_${TARGETARCH} ./cmd/podsh && \
  setcap 'cap_net_bind_service=+ep' podsh_${TARGETOS}_${TARGETARCH}

# Runner 
FROM scratch

COPY --from=runner-setup /etc/passwd /etc/group /etc/passwd
COPY --from=runner-setup /etc/group /etc/group

ENV HOME=/opt/podsh
ENV KUBECONFIG=/etc/podsh/kube/config

# Set up the working directory
WORKDIR /opt/podsh

ARG TARGETOS
ARG TARGETARCH

# Copy the binary from the builder stage
COPY --from=build-podsh --chmod=777 /app/podsh_${TARGETOS}_${TARGETARCH} /usr/bin/podsh

VOLUME /etc/podsh

EXPOSE 22
EXPOSE 8080

# Run the Go Gin binary
ENTRYPOINT ["/usr/bin/podsh"]


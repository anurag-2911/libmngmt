# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates (needed to fetch dependencies and for SSL)
RUN apk add --no-cache git ca-certificates tzdata

# Create appuser
ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Set working directory
WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o libmngmt \
    ./cmd/api

# Final stage - use distroless for better security
FROM gcr.io/distroless/static:nonroot

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy our static executable
COPY --from=builder /build/libmngmt /libmngmt

# Use an unprivileged user (nonroot user from distroless)
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/libmngmt"]

# Start from Alpine Linux image with the latest version of Golang
# Naming build stage as builder
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/ywadi/goq/
COPY . .
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
RUN go build -o /go/bin/crimsonq

FROM scratch
# Copy our static executable.
COPY --from=builder /go/bin/crimsonq /go/bin/crimsonq
# Run the hello binary.
ENTRYPOINT ["/go/bin/crimsonq"]
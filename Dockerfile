FROM golang:1.20.4 as builder

ARG COMMIT_SHA
ENV COMMIT_SHA=${COMMIT_SHA}

# Set the working directory
WORKDIR $GOPATH/src/github.com/diggerhq/cloud

# Copy all required source, blacklist files that are not required through `.dockerignore`
COPY . .

# Get the vendor library
RUN go version

# RUN vgo install

# https://github.com/ethereum/go-ethereum/issues/2738
# Build static binary "-getmode=vendor" does not work with go-ethereum
RUN go build
# -ldflags "-linkmode external -extldflags -static"

# Multi-stage build will just copy the binary to an alpine image.
FROM ubuntu:latest

WORKDIR /app

# Set gin to production
#ENV GIN_MODE=release

# Expose the running port
EXPOSE 3000

# Copy the binary to the corresponding folder
COPY --from=builder /go/src/github.com/diggerhq/cloud/cloud .

# Run the binary
CMD ["/app/cloud"]
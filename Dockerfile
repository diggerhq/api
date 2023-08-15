FROM golang:1.20.4 as builder
ARG COMMIT_SHA
RUN echo "commit sha: ${COMMIT_SHA}"

RUN update-ca-certificates

# Set the working directory
WORKDIR $GOPATH/src/github.com/diggerhq/cloud

# Copy all required source, blacklist files that are not required through `.dockerignore`
COPY . .

# Get the vendor library
RUN go version

# RUN vgo install

# https://github.com/ethereum/go-ethereum/issues/2738
# Build static binary "-getmode=vendor" does not work with go-ethereum
RUN go build -ldflags="-X 'main.Version=${COMMIT_SHA}'"

# Multi-stage build will just copy the binary to an alpine image.
FROM ubuntu:latest
ARG COMMIT_SHA
WORKDIR /app

RUN echo "commit sha: ${COMMIT_SHA}"

# Set gin to production
#ENV GIN_MODE=release

# Expose the running port
EXPOSE 3000

# Copy the binary to the corresponding folder
COPY --from=builder /go/src/github.com/diggerhq/cloud/cloud .
ADD templates ./templates

# Run the binary
CMD ["/app/cloud"]

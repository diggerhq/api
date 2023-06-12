FROM golang:1.10.3 as builder

# Set the working directory
WORKDIR $GOPATH/src/github.com/alextanhongpin/go-gin-starter

# Copy all required source, blacklist files that are not required through `.dockerignore`
COPY . .

# Get the vendor library
RUN go version && go get -v golang.org/x/vgo

# RUN vgo install

# https://github.com/ethereum/go-ethereum/issues/2738
# Build static binary "-getmode=vendor" does not work with go-ethereum
RUN go build -getmode=vendor -o app
# -ldflags "-linkmode external -extldflags -static"

# Multi-stage build will just copy the binary to an alpine image.
FROM alpine:3.7

RUN apk --no-cache add ca-certificates

WORKDIR /root

# Set gin to production
ENV GIN_MODE=release

# Expose the running port
EXPOSE 3000

# Copy the binary to the corresponding folder
COPY --from=builder /go/src/github.com/alextanhongpin/go-gin-starter/app .

ARG BUILD_DATE
ARG NAME
ARG VCS_URL
ARG VCS_REF
ARG VENDOR
ARG VERSION
ARG IMAGE_NAME

ENV BUILD_DATE $BUILD_DATE
ENV VERSION $VERSION

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name=$NAME \
      org.label-schema.description="go gin starter application" \
      org.label-schema.url="https://example.com" \
      org.label-schema.vcs-url=$VCS_URL \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vendor=$VENDOR \
      org.label-schema.version=$VERSION \
      org.label-schema.docker.schema-version="1.0" \
      org.label-schema.docker.cmd="docker run -d $IMAGE_NAME"

# Run the binary
CMD ["/root/app"]
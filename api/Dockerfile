####################
# Development image
####################
FROM golang:1.25-alpine AS development

# Environment variables
ENV CGO_ENABLED=0

VOLUME /src
WORKDIR /src

# Setup dependencies.
RUN apk --no-cache add ca-certificates git

# Install development dependencies.
RUN go get -u github.com/cespare/reflex

# Download dependencies into the image prior to copying over any source files.
# This allows the downloading dependencies step to be cached in a layer so that
# we don't have to repeat the download each time a source file changes.
COPY go.mod .
COPY go.sum .
RUN go get ./...

# Copy all of the source code into the image so that it can be compiled.  When
# running via docker-compose in development mode we'll override this and instead
# mount the source directory from the host machine.
COPY . .

# Compile into a static binary.
RUN go get && \
    go build -o /api


####################
# Production image
####################
FROM scratch

EXPOSE 5000

# Copy over the static executable and certs (needed because we call into http
# APIs).
COPY --from=development /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=development /api /api

ENTRYPOINT ["/api"]

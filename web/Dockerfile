####################
# Development image
####################
FROM golang:1.13-alpine AS development

# The environment to use for assets.
ARG ENVIRONMENT=prod

# Setup dependencies
RUN apk --no-cache add ca-certificates git

# Install development dependencies
RUN go get -u                              \
      github.com/kevinburke/go-bindata/... \
      github.com/cespare/reflex

# Copy all of the source code into the image so that it can be compiled.  When
# running via docker-compose in development mode we'll override this and instead
# mount the source directory from the host machine.
VOLUME /src
WORKDIR /src
COPY . .

# Compile into a static binary.
ENV CGO_ENABLED=0
ENV GO111MODULE=on
RUN go generate -tags ${ENVIRONMENT} ./... && \
    go get                                 && \
    go build -tags ${ENVIRONMENT} -o /web


####################
# Production image
####################
FROM scratch

EXPOSE 5000

# Copy over the static executable and certs (needed because we call into http
# APIs).
COPY --from=development /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=development /web /web

ENV GIN_MODE=release
ENTRYPOINT ["/web"]

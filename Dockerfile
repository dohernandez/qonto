# --- BEGINING OF BUILDER

FROM golang:1.17.3 AS builder

WORKDIR /go/src/github.com/dohernandez/qonto

# This is to cache the Go modules in their own Docker layer by
# using `go mod download`, so that next steps in the Docker build process
# won't need to download modules again if no modules have been updated.
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Install migrate
RUN  curl -sL https://github.com/golang-migrate/migrate/releases/download/v4.11.0/migrate.linux-amd64.tar.gz | tar xvz \
    && mv migrate.linux-amd64 /bin/migrate

COPY . ./

# Build http binary and cli binary
RUN make build

# --- END OF BUILDER

FROM ubuntu:focal

RUN groupadd -r gonto && useradd --no-log-init -r -g gonto gonto
USER sportbuf

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder --chown=gonto:gonto /go/src/github.com/dohernandez/qonto/bin/qonto /bin/qonto
COPY --from=builder --chown=gonto:gonto /go/src/github.com/dohernandez/qonto/resources/migrations /resources/migrations
COPY --from=builder --chown=gonto:gonto /bin/migrate /bin/migrate

EXPOSE 8000 8080 8010
ENTRYPOINT ["gonto"]

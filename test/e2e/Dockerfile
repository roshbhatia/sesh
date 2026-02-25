FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/sesh .

# ── runtime image ────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache git bash

COPY --from=builder /usr/local/bin/sesh /usr/local/bin/sesh
COPY test/e2e/run_e2e.sh /run_e2e.sh
RUN chmod +x /run_e2e.sh

ENTRYPOINT ["/run_e2e.sh"]

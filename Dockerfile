FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata make curl

RUN adduser -D -g '' appuser

WORKDIR /build

COPY . .

RUN make ci-setup

RUN make build

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /build/myip /myip
COPY --from=builder /usr/bin/curl /usr/bin/curl

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/usr/bin/curl", "-f", "http://localhost:8080/health"]

ENTRYPOINT ["/myip"]

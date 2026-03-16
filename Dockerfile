FROM golang:1 AS builder
COPY . /app
WORKDIR /app
RUN make build

FROM gcr.io/distroless/static-debian13:latest
COPY --from=builder /app/bin/ecosystem-activity /ecosystem-activity
COPY config.yaml /config.yaml
USER nonroot:nonroot
CMD ["/ecosystem-activity"]

FROM golang:1.22 as builder
COPY main.go /main.go
ENV GOROOT=/usr/local/go
RUN CGO_ENABLED=0 go build -o /wisebalance /main.go

FROM gcr.io/distroless/base
COPY --from=builder /wisebalance /wisebalance
USER 65534:65534
ENTRYPOINT ["/wisebalance"]

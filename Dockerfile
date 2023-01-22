FROM golang:1.19 as builder

WORKDIR /cov-diff
COPY . /cov-diff

RUN go mod download

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o cov-diff cmd/cov-div/main.go

FROM gcr.io/distroless/static

COPY --from=builder /cov-diff/cov-diff /cov-diff

ENTRYPOINT ["/cov-diff"]

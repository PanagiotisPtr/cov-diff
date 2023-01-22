FROM golang:1.19-alpine

WORKDIR /cov-diff
COPY . /cov-diff

RUN go mod download

RUN cd cmd/cov-diff && go install

CMD ["cov-diff"]

FROM golang:1.21

WORKDIR /cov-diff
COPY . /cov-diff

RUN go mod download

RUN cd cmd/cov-diff && go install

CMD ["cov-diff"]

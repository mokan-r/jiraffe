FROM alpine

WORKDIR /jiraffe/

COPY . .

RUN apk add --no-cache musl-dev go make

ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin && \
    go mod tidy && \
    go build github.com/mokan-r/jiraffe/cmd/bot

FROM alpine

WORKDIR /jiraffe/

COPY bot .

ENTRYPOINT ["./bot"]

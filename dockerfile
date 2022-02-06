#Build 
FROM golang:alpine AS builder
LABEL stage=builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/ywadi/crimsonq/
COPY . .
RUN go get -d -v
RUN go build -o /go/bin/crimsonq
#Run
FROM alpine
RUN apk update
COPY --from=builder /go/bin/crimsonq /go/bin/crimsonq
RUN mkdir -p /CrimsonQ/.crimsonQ
WORKDIR /CrimsonQ/.crimsonQ
COPY ./crimson.config /CrimsonQ/.crimsonQ
EXPOSE 9001

ENTRYPOINT ["/go/bin/crimsonq"]
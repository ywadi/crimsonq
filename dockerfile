#Build CrimsonQ DB & Service 
FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/ywadi/crimsonq/
COPY . .
RUN go get -d -v
RUN go build -o /go/bin/crimsonq
WORKDIR $GOPATH/src/ywadi/crimsonq/cmd
RUN go get -d -v
RUN go build -o /go/bin/crimsonq-cli

#Build CrimsonQ Dashboard 
FROM node:12-alpine3.14 AS dashbuilder
RUN apk update && apk add --no-cache git
WORKDIR /
RUN git clone https://github.com/Ola-Alkhateeb/crimsonQ-dashboard.git
WORKDIR /crimsonQ-dashboard
RUN echo 'VUE_APP_API_URL="/api/"' >> .env
RUN echo  'VUE_APP_LOGIN_URL="../login/"' >> .env 
RUN npm install
RUN npm run build 

#Run
FROM alpine
RUN apk update
COPY --from=builder /go/bin/crimsonq /go/bin/crimsonq
COPY --from=builder /go/bin/crimsonq-cli /bin/crimsonq-cli
COPY --from=dashbuilder /crimsonQ-dashboard/dist/ /WebUI/
RUN mkdir -p /CrimsonQ/.crimsonQ
WORKDIR /CrimsonQ/.crimsonQ
COPY ./crimson.config /
EXPOSE 9001
EXPOSE 8080

ENTRYPOINT ["/go/bin/crimsonq"]
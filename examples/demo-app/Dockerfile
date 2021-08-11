FROM golang:1.16.7
WORKDIR /usr/src/demo-app
COPY . .
RUN go install ./main.go
CMD ["/go/bin/main"]

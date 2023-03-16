FROM golang:1.20 as builder
RUN update-ca-certificates
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-extldflags=-static' -o guestbook *.go

FROM scratch
LABEL maintainer="Patrick D'appollonio <hey@patrickdap.com>"
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/guestbook /app/guestbook
EXPOSE 80
CMD ["./app/guestbook"]

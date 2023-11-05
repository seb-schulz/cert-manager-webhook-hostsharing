FROM docker.io/golang:1.19-alpine

RUN apk add --no-cache ca-certificates git make curl build-base bash && \
    go install -v golang.org/x/tools/gopls@latest

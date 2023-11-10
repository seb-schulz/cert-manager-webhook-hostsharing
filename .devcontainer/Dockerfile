ARG GO_TAG=1.21-alpine

FROM docker.io/golang:${GO_TAG}

RUN apk add --no-cache ca-certificates git make curl build-base bash \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install golang.org/x/tools/gopls@latest \
    && go install -v golang.org/x/tools/cmd/goimports@latest

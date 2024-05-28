FROM docker.io/golang:1.22-alpine

RUN apk add --no-cache ca-certificates git make curl build-base bash sudo shadow \
    && groupadd --gid 1000 vscode \
    && useradd --uid 1000 --gid 1000 -m vscode \
    && echo vscode ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/vscode \
    && chmod 0440 /etc/sudoers.d/vscode

USER vscode

RUN go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install golang.org/x/tools/gopls@latest \
    && go install -v golang.org/x/tools/cmd/goimports@latest

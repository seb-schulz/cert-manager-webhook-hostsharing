ARG GO_TAG=1.23-alpine@sha256:ac67716dd016429be8d4c2c53a248d7bcdf06d34127d3dc451bda6aa5a87bc06
ARG ALPINE_TAG=3.20@sha256:beefdbd8a1da6d2915566fde36db9db0b524eb737fc57cd1367effd16dc0d06d

FROM docker.io/golang:${GO_TAG} AS build_deps

WORKDIR /workspace

COPY go.mod .
COPY go.sum .

RUN go mod download

FROM build_deps AS build

COPY . .

RUN CGO_ENABLED=0 go build -o webhook -ldflags '-w -extldflags "-static"' ./cmd/webhook/
RUN CGO_ENABLED=0 go build -o _out/updater -ldflags '-w -extldflags "-static"' ./cmd/updater/

FROM docker.io/alpine:${ALPINE_TAG}

RUN apk add --no-cache ca-certificates

COPY --from=build /workspace/webhook /usr/local/bin/webhook

ENTRYPOINT ["webhook"]

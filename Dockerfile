ARG GO_TAG=1.23-alpine@sha256:c694a4d291a13a9f9d94933395673494fc2cc9d4777b85df3a7e70b3492d3574
ARG ALPINE_TAG=3.20@sha256:1e42bbe2508154c9126d48c2b8a75420c3544343bf86fd041fb7527e017a4b4a

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

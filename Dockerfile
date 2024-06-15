ARG GO_TAG=1.22-alpine@sha256:6522f0ca555a7b14c46a2c9f50b86604a234cdc72452bf6a268cae6461d9000b
ARG ALPINE_TAG=3.20@sha256:77726ef6b57ddf65bb551896826ec38bc3e53f75cdde31354fbffb4f25238ebd

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

ARG GO_TAG=1.23-alpine@sha256:6c5c9590f169f77c8046e45c611d3b28fe477789acd8d3762d23d4744de69812
ARG ALPINE_TAG=3.21@sha256:21dc6063fd678b478f57c0e13f47560d0ea4eeba26dfc947b2a4f81f686b9f45

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

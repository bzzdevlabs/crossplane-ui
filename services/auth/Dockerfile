# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ENV CGO_ENABLED=0
ENV GOFLAGS=-trimpath
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build \
        -ldflags "-s -w \
            -X gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/auth/internal/buildinfo.Version=${VERSION} \
            -X gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/auth/internal/buildinfo.Commit=${COMMIT} \
            -X gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/auth/internal/buildinfo.Date=${BUILD_DATE}" \
        -o /out/auth ./cmd

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

LABEL org.opencontainers.image.title="crossplane-ui auth"
LABEL org.opencontainers.image.source="https://gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui"
LABEL org.opencontainers.image.licenses="Apache-2.0"

USER 65532:65532
WORKDIR /
COPY --from=builder /out/auth /usr/local/bin/auth

EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/auth"]

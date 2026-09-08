# Build stage: compile a static binary
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/status-api ./cmd/status-api \
 && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/statusctl ./cmd/statusctl

# Run stage: distroless, no shell, runs as non root
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/status-api /usr/local/bin/status-api
COPY --from=build /out/statusctl /usr/local/bin/statusctl
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/status-api"]

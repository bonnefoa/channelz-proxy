FROM golang:latest as builder

ARG GIT_COMMIT
ARG GIT_BRANCH
ARG GO_VERSION
ARG VERSION
ARG BUILD_DATE

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on \
    go build -a -o channelz-proxy \
        -ldflags "-X 'main.gitCommit=$GIT_COMMIT' -X 'main.gitBranch=$GIT_BRANCH' -X 'main.goVersion=$GO_VERSION' -X 'main.buildDate=$BUILD_DATE' -X 'main.version=$VERSION'" \
        cmd/main.go

FROM alpine:latest
WORKDIR /
COPY --from=builder /workspace/channelz-proxy .

ENTRYPOINT ["/channelz-proxy"]

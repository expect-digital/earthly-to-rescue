VERSION 0.8
PROJECT expect.digital/demo
ARG --global go_version=1.21.11
FROM golang:$go_version-alpine
WORKDIR /counter

# up installs the project in local docker.
up:
  LOCALLY
  WITH DOCKER --load=+image
    RUN docker compose up --force-recreate -d
  END

# down uninstalls the project in local docker.
down:
  LOCALLY
  RUN docker compose down --remove-orphans

all:
  BUILD +check
  BUILD +image

deps:
  COPY go.mod go.sum .
  RUN \
    --mount=type=cache,target=/go/pkg/mod,id=go-mod \
    --mount=type=cache,target=/root/.cache/go-build,id=go-build \
    go mod download
  RUN go mod edit -go=$go_version

go:
  FROM +deps
  COPY *.go .
  SAVE ARTIFACT /counter

# build compiles the counter CLI and saves it as bin/counter.
build:
  FROM +go
  ARG GOOS=linux
  ARG GOARCH=arm64
  ARG output=bin/counter.$GOOS.$GOARCH
  RUN \
    --mount=type=cache,target=/go/pkg/mod,id=go-mod \
    --mount=type=cache,target=/root/.cache/go-build,id=go-build \
    go build -o bin/counter
  SAVE ARTIFACT bin/counter counter AS LOCAL $output

# image builds an image for creating a container.
image:
  ARG tag=latest
  FROM alpine
  COPY +build/counter .
  ENTRYPOINT ["/counter"]
  SAVE IMAGE counter/counter:$tag

# check verifies code quality by running linters and tests.
check:
  BUILD +test
  BUILD +lint

# test runs unit and integration tests.
test:
  BUILD +test-unit
  BUILD +test-integration

# test-unit runs unit tests.
test-unit:
  FROM +go
  RUN \
    --mount=type=cache,target=/go/pkg/mod,id=go-mod \
    --mount=type=cache,target=/root/.cache/go-build,id=go-build \
    go test ./...

# test-integration runs integration tests.
test-integration:
  FROM earthly/dind:alpine-3.20-docker-26.1.3-r0
  COPY compose.yaml .
  COPY --dir +go/counter /
  WITH DOCKER --compose compose.yaml --service redis --pull golang:$go_version-alpine
    RUN \
    --mount=type=cache,target=/go/pkg/mod,id=go-mod \
    --mount=type=cache,target=/root/.cache/go-build,id=go-build \
    docker run \
      --network=host \
      -v /go/pkg/mod:/go/pkg/mod \
      -v /root/.cache/go-build:/root/.cache/go-build \
      -v /counter:/counter \
      -e REDIS_ADDR=localhost:6379 \
      golang:$go_version-alpine go test -C /counter -tags=integration ./...
  END

# lint analyses code for errors, bugs and stylistic issues (golangci-lint).
lint:
  ARG golangci_lint_version=1.59.0
  RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /go/bin v$golangci_lint_version
  COPY --dir +go/counter /
  COPY .golangci.yaml .
  RUN \
    --mount=type=cache,target=/go/pkg/mod,id=go-mod \
    --mount=type=cache,target=/root/.cache/go-build,id=go-build \
    --mount=type=cache,target=/root/.cache/golangci-lint \
    golangci-lint run
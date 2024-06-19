package main

import (
	"context"
	"fmt"

	"go.expect.digital/counter/dagger/internal/dagger"
	"golang.org/x/sync/errgroup"
)

func cache(key string) *dagger.CacheVolume { return dag.CacheVolume(key) }

type DaggerToRescue struct {
	GoVersion string
}

func New(
	// +optional
	// +default="1.21.11"
	goVersion string,
) *DaggerToRescue {
	return &DaggerToRescue{
		GoVersion: goVersion,
	}
}

func (m *DaggerToRescue) src(src *Directory) *Container {
	return dag.Container().
		From(fmt.Sprintf("golang:%s-alpine", m.GoVersion)).
		WithMountedCache("/go/pkg/mod", cache("gomod")).
		WithMountedCache("/root/.cache/go-build", cache("gobuild")).
		WithWorkdir("/counter").
		WithDirectory(".", src).
		WithExec([]string{"go", "mod", "download"}).
		WithExec([]string{"go", "mod", "edit", "-go=" + m.GoVersion})
}

// Compiles the counter.
func (m *DaggerToRescue) Build(
	src *Directory,
	// +optional
	// +default="linux"
	os string,
	// +optional
	// +default="arm64"
	arch string,
) *Directory {
	return dag.Container().
		From(fmt.Sprintf("golang:%s-alpine", m.GoVersion)).
		WithMountedCache("/go/pkg/mod", cache("gomod")).
		WithMountedCache("/root/.cache/go-build", cache("gobuild")).
		WithEnvVariable("GOOS", os).
		WithEnvVariable("GOARCH", arch).
		WithWorkdir("/counter").
		WithDirectory("/counter", src).
		WithExec([]string{"go", "build", "-o", "bin/counter"}).
		Directory("bin")
}

// Builds an image for creating a container.
func (m *DaggerToRescue) Image(
	ctx context.Context,
	src *Directory,
	// +optional
	// +default="linux"
	os string,
	// +optional
	// +default="arm64"
	arch string,
	// +optional
	// +default="latest"
	tag string,
) (string, error) {
	bin := m.Build(src, os, arch)

	return dag.Container(ContainerOpts{Platform: Platform(os + "/" + arch)}).
		From("alpine").
		WithFile(".", bin.File("counter")).
		WithEntrypoint([]string{"/counter"}).
		Publish(ctx, "counter/counter:"+tag)
}

// Runs unit tests.
func (m *DaggerToRescue) TestUnit(ctx context.Context, src *Directory) (string, error) {
	return m.src(src).
		WithMountedCache("/go/pkg/mod", cache("gomod")).
		WithMountedCache("/root/.cache/go-build", cache("gobuild")).
		WithExec([]string{"go", "test", "./..."}).
		Stdout(ctx)
}

// Runs integration tests.
func (m *DaggerToRescue) TestIntegration(ctx context.Context, src *Directory) (string, error) {
	redis := dag.Container().
		From("redis:7.0.15-alpine").
		WithExposedPort(6379). //nolint:mnd
		AsService()

	return m.src(src).
		WithServiceBinding("redis", redis).
		WithEnvVariable("REDIS_ADDR", "redis:6379").
		WithMountedCache("/go/pkg/mod", cache("gomod")).
		WithMountedCache("/root/.cache/go-build", cache("gobuild")).
		WithExec([]string{"go", "test", "-tags=integration", "./..."}).
		Stdout(ctx)
}

// Runs unit and integration tests.
func (m *DaggerToRescue) Test(ctx context.Context, src *Directory) (string, error) {
	eg, egCtx := errgroup.WithContext(ctx)
	out := [2]string{}

	eg.Go(func() error {
		var err error

		out[0], err = m.TestUnit(egCtx, src)

		return err
	})

	eg.Go(func() error {
		var err error

		out[1], err = m.TestIntegration(egCtx, src)

		return err
	})

	err := eg.Wait()

	return out[0] + out[1], err
}

// Analyses code for errors, bugs and stylistic issues (golangci-lint).
func (m *DaggerToRescue) Lint(
	ctx context.Context,
	src *Directory,
	// +optional
	// +default="1.59.0"
	golangciLintVersion string,
) (string, error) {
	dir := m.src(src).Directory("/counter")

	return dag.Container().
		From(fmt.Sprintf("golang:%s-alpine", m.GoVersion)).
		WithExec([]string{"sh", "-c", "wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /go/bin v" + golangciLintVersion}). //nolint:lll
		WithWorkdir("/counter").
		WithDirectory(".", dir).
		WithMountedCache("/go/pkg/mod", cache("gomod")).
		WithMountedCache("/root/.cache/go-build", cache("gobuild")).
		WithMountedCache("/root/.cache/golangci-lint", cache("golangci")).
		WithExec([]string{"golangci-lint", "run"}).
		Stdout(ctx)
}

// Verifies code quality by running linters and tests.
func (m *DaggerToRescue) Check(
	ctx context.Context,
	src *Directory,
	// +optional
	// +default="1.59.0"
	golangciLintVersion string,
) (string, error) {
	eg, egCtx := errgroup.WithContext(ctx)
	out := [2]string{}

	eg.Go(func() error {
		var err error

		out[0], err = m.Test(egCtx, src)

		return err
	})

	eg.Go(func() error {
		var err error

		out[1], err = m.Lint(egCtx, src, golangciLintVersion)

		return err
	})

	err := eg.Wait()

	return out[0] + out[1], err
}

func (m *DaggerToRescue) All(
	ctx context.Context,
	src *Directory,
	// +optional
	// +default="linux"
	os string,
	// +optional
	// +default="arm64"
	arch string,
	// +optional
	// +default="latest"
	tag string,
	// +optional
	// +default="1.59.0"
	golangciLintVersion string,
) (string, error) {
	var out [2]string

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var err error

		out[0], err = m.Image(egCtx, src, os, arch, tag)

		return err
	})

	eg.Go(func() error {
		var err error

		out[1], err = m.Check(egCtx, src, golangciLintVersion)

		return err
	})

	err := eg.Wait()

	return out[0] + out[1], err
}

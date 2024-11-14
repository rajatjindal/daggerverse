package main

import (
	"context"
	"dagger/wasi/internal/dagger"
	"fmt"
	"strings"
)

type Wasi struct {
	BaseImage        string
	GolangVersion    string
	TinygoVersion    string
	RustVersion      string
	WasmtoolsVersion string
	SpinVersion      string
}

func New(
	// The base image to use for the build env.
	//
	// +default="ubuntu:24.04"
	baseImage string,

	// The golang version to be installed for the build env.
	//
	// This version should be compatible with the tinygo version used.
	//
	// +default="1.23.2"
	golangVersion string,

	// The tinygo version to be installed for the build env.
	//
	// This version should be compatible with the golang version used.
	//
	// +default="0.34.0"
	tinygoVersion string,

	// The rust version to be installed for the build env.
	//
	// +default="1.82.0"
	rustVersion string,

	// The version of wasm-tools to be installed for build env.
	//
	// This is required when developing the golang based app.
	//
	// +default="1.220.0"
	wasmtoolsVersion string,

	// The version of spin to be installed for build env.
	//
	// +default="3.0.0"
	spinVersion string,
) *Wasi {
	// panic(fmt.Sprintf("inside new %s", wasmtoolsVersion))
	return &Wasi{
		BaseImage:        baseImage,
		GolangVersion:    golangVersion,
		TinygoVersion:    tinygoVersion,
		RustVersion:      rustVersion,
		WasmtoolsVersion: wasmtoolsVersion,
		SpinVersion:      spinVersion,
	}
}

func (w *Wasi) Base() *dagger.Container {
	return dag.Container().
		From(w.BaseImage).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "wget"})
}

func (w *Wasi) BuildEnv(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) (*dagger.Container, error) {
	ctr := w.Base().
		WithWorkdir("/app").
		WithDirectory("/app", source)

	var toolchains []string
	projectToolchains, err := ctr.File(".toolchains").Contents(ctx)
	if err == nil && projectToolchains != "" {
		toolchains = strings.Split(projectToolchains, "\n")
	}

	// change workdir to /tmp while we install toolchains
	ctr = ctr.WithWorkdir("/tmp/")

	var installedToolchains = map[string]string{}
	for _, toolchain := range toolchains {
		if toolchain == "" {
			continue
		}

		name, version := getToolchainVersion(toolchain, w.GolangVersion)
		withFunc, exists := withToolchainMap[name]
		if !exists {
			return nil, fmt.Errorf("unknown toolchain requested %q", name)
		}

		ctr = ctr.With(withFunc(version))
		installedToolchains[name] = version
	}

	// ensure spin is always installed
	if _, ok := installedToolchains["spin"]; !ok {
		ctr = ctr.With(WithSpin(w.SpinVersion))
	}

	// change workdir back to /app
	ctr = ctr.WithWorkdir("/app")
	return ctr, nil
}

func (w *Wasi) Build(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default=[]
	args []string,
) (*dagger.Container, error) {
	buildctr, err := w.BuildEnv(ctx, source)
	if err != nil {
		return nil, err
	}

	return buildctr.WithExec(append([]string{"spin", "build"}, args...), dagger.ContainerWithExecOpts{Expect: dagger.Any}).Sync(ctx)
}

func (w *Wasi) RegistryPush(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	ociArtifactName string,
	// +default=[]
	args []string,
) (*dagger.Container, error) {
	buildctr, err := w.BuildEnv(ctx, source)
	if err != nil {
		return nil, err
	}

	return buildctr.
		WithExec(append([]string{"spin", "registry", "push", ociArtifactName}, args...), dagger.ContainerWithExecOpts{Expect: dagger.Any}).
		Sync(ctx)
}

func getToolchainVersion(toolchain, defaultVersion string) (string, string) {
	if !strings.Contains(toolchain, "=") {
		return toolchain, defaultVersion
	}

	parts := strings.Split(toolchain, "=")

	return parts[0], parts[1]
}

var withToolchainMap = map[string]func(version string) dagger.WithContainerFunc{
	"go":     WithGoToolchain,
	"golang": WithGoToolchain,
	"rust":   WithRustToolchain,
	"tinygo": WithTinyGoToolchain,
	"spin":   WithSpin,
}

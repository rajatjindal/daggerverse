package main

import (
	"context"
	"strings"

	"dagger/toolchains/internal/dagger"
)

type Toolchains struct {
	Golang     string
	Postgresql string
}

func New(
	// +default="1.23.6"
	golang string,

	// +default="17.4"
	postgresql string,
) *Toolchains {
	return &Toolchains{
		Golang:     golang,
		Postgresql: postgresql,
	}
}

func (t *Toolchains) GetRequiredVersions(ctx context.Context, source *dagger.Directory) (map[string]string, error) {
	projectToolchains, err := source.File(".toolchains").Contents(ctx)
	if err != nil {
		return nil, err
	}

	output := map[string]string{}
	toolchains := strings.Split(projectToolchains, "\n")
	for _, toolchain := range toolchains {
		name, version := getToolchainVersion(toolchain, t.Golang)
		output[name] = version
	}

	return output, nil
}

func getToolchainVersion(toolchain, defaultVersion string) (string, string) {
	if !strings.Contains(toolchain, "=") {
		return toolchain, defaultVersion
	}

	parts := strings.Split(toolchain, "=")

	return parts[0], parts[1]
}

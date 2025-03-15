package main

import (
	"context"
	"dagger/toolchains/internal/dagger"
	"strings"
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

func (t *Toolchains) InitRequiredVersions(ctx context.Context, source *dagger.Directory) (*Toolchains, error) {
	projectToolchains, err := source.File(".toolchains").Contents(ctx)
	if err != nil {
		return nil, err
	}

	toolchains := strings.Split(projectToolchains, "\n")
	for _, toolchain := range toolchains {
		name, version := getToolchainVersion(toolchain, t.Golang)
		switch name {
		case "golang", "go":
			t.Golang = version
		case "postgres", "postgresql":
			t.Postgresql = version
		}
	}

	return t, nil
}

func getToolchainVersion(toolchain, defaultVersion string) (string, string) {
	if !strings.Contains(toolchain, "=") {
		return toolchain, defaultVersion
	}

	parts := strings.Split(toolchain, "=")

	return parts[0], parts[1]
}

package main

import (
	"dagger/wasi/internal/dagger"
	"fmt"
	"path/filepath"
	"runtime"
)

func WithRustToolchain(version string) dagger.WithContainerFunc {
	return func(c *dagger.Container) *dagger.Container {
		return c.
			WithExec([]string{"sh", "-c", "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y"}).
			WithEnvVariable("PATH", "$PATH:/root/.cargo/bin", dagger.ContainerWithEnvVariableOpts{Expand: true}).
			WithExec([]string{"rustup", "toolchain", "install", version, "--component=clippy", "--component=rustfmt", "--no-self-update"}).
			WithExec([]string{"rustup", "default", version}).
			WithExec([]string{"rustup", "target", "add", "wasm32-wasi"}).
			WithExec([]string{"rustup", "target", "add", "wasm32-unknown-unknown"})
	}
}

func WithGoToolchain(version string) dagger.WithContainerFunc {
	return func(c *dagger.Container) *dagger.Container {
		releaseArtifactName := fmt.Sprintf("go%s.%s-%s", version, runtime.GOOS, runtime.GOARCH)
		releaseArtifactTarFile := fmt.Sprintf("%s.tar.gz", releaseArtifactName)
		releaseArtifactDownloadLink := fmt.Sprintf("https://go.dev/dl/%s", releaseArtifactTarFile)
		return c.
			WithExec([]string{"wget", releaseArtifactDownloadLink, "-O", releaseArtifactTarFile}).
			WithExec([]string{"rm", "-rf", "/usr/local/go"}).
			WithExec([]string{"tar", "-C", "/usr/local", "-xvf", releaseArtifactTarFile}).
			WithEnvVariable("PATH", "/usr/local/go/bin:$PATH", dagger.ContainerWithEnvVariableOpts{
				Expand: true,
			})
	}
}

func WithTinyGoToolchain(version string) dagger.WithContainerFunc {
	return func(c *dagger.Container) *dagger.Container {
		releaseArtifactName := fmt.Sprintf("tinygo%s.%s-%s", version, runtime.GOOS, runtime.GOARCH)
		releaseArtifactTarFile := fmt.Sprintf("%s.tar.gz", releaseArtifactName)
		releaseArtifactDownloadLink := fmt.Sprintf("https://github.com/tinygo-org/tinygo/releases/download/v%s/%s", version, releaseArtifactTarFile)
		return c.
			WithExec([]string{"wget", releaseArtifactDownloadLink, "-O", releaseArtifactTarFile}).
			WithExec([]string{"tar", "-xvf", releaseArtifactTarFile}).
			WithExec([]string{"mkdir", "-p", "/opt"}).
			WithExec([]string{"mv", "tinygo", "/opt/tinygo"}).
			WithEnvVariable("TINYGOROOT", "/opt/tinygo").
			WithEnvVariable("PATH", "/opt/tinygo/bin:$PATH", dagger.ContainerWithEnvVariableOpts{
				Expand: true,
			})
	}
}

func WithWasmTools(version string) dagger.WithContainerFunc {
	return func(c *dagger.Container) *dagger.Container {
		arch := "x86_64"
		if runtime.GOARCH == "arm64" {
			arch = "aarch64"
		}

		releaseArtifactName := fmt.Sprintf("wasm-tools-%s-%s-%s", version, arch, runtime.GOOS)
		releaseArtifactTarFile := fmt.Sprintf("%s.tar.gz", releaseArtifactName)
		releaseArtifactDownloadLink := fmt.Sprintf("https://github.com/bytecodealliance/wasm-tools/releases/download/v%s/%s", version, releaseArtifactTarFile)
		return c.
			WithExec([]string{"wget", releaseArtifactDownloadLink, "-O", releaseArtifactTarFile}).
			WithExec([]string{"tar", "-xvf", releaseArtifactTarFile}).
			WithExec([]string{"mv", filepath.Join(releaseArtifactName, "wasm-tools"), "/usr/local/bin/wasm-tools"})
	}
}

func WithSpin(version string) dagger.WithContainerFunc {
	return func(c *dagger.Container) *dagger.Container {
		// TODO(rajatjindal): allow pulling specific version
		// right now defaults to canary
		return c.WithFile("/usr/local/bin/spin", dag.Container().
			From("ghcr.io/fermyon/spin:canary-distroless").File("/usr/local/bin/spin"))
	}
}

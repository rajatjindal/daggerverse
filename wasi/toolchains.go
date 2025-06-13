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
			WithExec([]string{"rustup", "target", "add", "wasm32-wasip1"}).
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
			}).
			// TO GET WIT FILES
			WithDirectory("/tmp/really-tmp", dag.Directory()).
			WithWorkdir("/tmp/really-tmp").
			WithExec([]string{"go", "mod", "init", "really-tmp"}).
			WithExec([]string{"go", "get", "-u", "github.com/spinframework/spin-go-sdk/v2@wasip2"}).
			WithoutDirectory("/tmp/really-tmp")
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

func WithNode(version string) dagger.WithContainerFunc {
	return func(c *dagger.Container) *dagger.Container {
		return c.
			WithExec([]string{"wget", "https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.5/install.sh", "-O", "install.sh"}).
			WithExec([]string{"bash", "install.sh"}).
			WithEnvVariable("NVM_DIR", "/root/.nvm", dagger.ContainerWithEnvVariableOpts{Expand: true}).
			WithExec([]string{"sh", "-c", ". /root/.nvm/nvm.sh && nvm --version"}).
			WithExec([]string{"sh", "-c", fmt.Sprintf(". /root/.nvm/nvm.sh && nvm install %s", version)}).
			WithExec([]string{"sh", "-c", fmt.Sprintf(". /root/.nvm/nvm.sh && nvm use %s", version)}).
			WithEnvVariable("PATH", fmt.Sprintf("/root/.nvm/versions/node/v%s/bin:$PATH", version), dagger.ContainerWithEnvVariableOpts{Expand: true}).
			WithExec([]string{"npm", "install", "-g", "yarn"})
	}
}

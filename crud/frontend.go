package main

import (
	"context"
	"dagger/dev/internal/dagger"
	"fmt"
)

const (
	nodeVersion = "20.17.0"
)

type Frontend struct {
	Dev *Dev
	Src *dagger.Directory
}

func (m *Frontend) Generate(ctx context.Context) *dagger.Directory {
	return dag.Container().
		From(fmt.Sprintf("node:%s-alpine", nodeVersion)).
		WithMountedCache("/usr/local/share/.cache/yarn", dag.CacheVolume("global-yarn-cache")).
		WithMountedDirectory("/work", m.Src).
		WithWorkdir("/work").
		WithExec([]string{"yarn", "install"}).
		WithExec([]string{"yarn", "generate", "--dotenv", ".env.local"}).
		Directory(".output/public")
}

func (m *Frontend) Build(ctx context.Context) *dagger.Container {
	return dag.
		Container().
		From(fmt.Sprintf("node:%s-alpine", nodeVersion)).
		WithMountedCache("/usr/local/share/.cache/yarn", dag.CacheVolume("global-yarn-cache")).
		WithDirectory("/app", m.Generate(ctx)).
		WithWorkdir("/app").
		WithExec([]string{"yarn", "add", "serve"}).
		WithExposedPort(3000).
		WithExec([]string{"npx", "serve", ".", "--debug", "--cors", "--no-port-switching", "-l", "3002"})
}

func (m *Frontend) Serve(ctx context.Context) *dagger.Service {
	return m.Build(ctx).AsService()
}

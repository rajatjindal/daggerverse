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
	Crud *Crud
	Src  *dagger.Directory
}

func (m *Frontend) Generate(ctx context.Context) *dagger.Directory {
	return dag.Container().
		From(fmt.Sprintf("node:%s-alpine", nodeVersion)).
		WithExec([]string{"npm", "install", "-g", fmt.Sprintf("pnpm@%s", "9.6.0")}).
		WithMountedCache("/root/.pnpm-store", dag.CacheVolume("pnpm-cache")).
		// WithMountedCache("/usr/local/share/.cache/yarn", dag.CacheVolume("global-yarn-cache")).
		WithMountedDirectory("/work", m.Src).
		WithoutDirectory("/work/node_modules").
		WithWorkdir("/work").
		WithExec([]string{"pnpm", "install"}).
		WithExec([]string{"pnpm", "generate", "--dotenv", ".env.local"}).
		Directory(".output/public")
}

func (m *Frontend) Build(ctx context.Context) *dagger.Container {
	return dag.
		Container().
		From(fmt.Sprintf("node:%s-alpine", nodeVersion)).
		WithExec([]string{"npm", "install", "-g", fmt.Sprintf("pnpm@%s", "9.6.0")}).
		WithMountedCache("/root/.pnpm-store", dag.CacheVolume("pnpm-cache")).
		// WithMountedCache("/usr/local/share/.cache/yarn", dag.CacheVolume("global-yarn-cache")).
		WithDirectory("/app", m.Generate(ctx)).
		WithWorkdir("/app").
		WithExec([]string{"pnpm", "add", "serve"}).
		WithExposedPort(3000)
}

func (m *Frontend) Serve(ctx context.Context) *dagger.Service {
	return m.Build(ctx).AsService(dagger.ContainerAsServiceOpts{
		Args: []string{"npx", "serve", ".", "--debug", "--cors", "--no-port-switching", "-l", "3000"}},
	)
}

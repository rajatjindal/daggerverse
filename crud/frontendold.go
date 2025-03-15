package main

import (
	"context"
	"dagger/dev/internal/dagger"
	"fmt"
)

const (
	oldNodeVersion = "16.13.1"
)

type FrontendOld struct {
	Crud *Crud
	Src  *dagger.Directory
}

func (m *FrontendOld) Build(ctx context.Context) *dagger.Container {
	return dag.Container().
		From(fmt.Sprintf("node:%s-alpine", oldNodeVersion)).
		WithMountedCache("/usr/local/share/.cache/yarn", dag.CacheVolume("global-yarn-cache")).
		WithMountedDirectory("/work", m.Src).
		WithWorkdir("/work").
		WithExec([]string{"yarn", "install"}).
		WithExec([]string{"sh", "-c", `sed -i "s/import nanoid from 'nanoid'/import { nanoid } from 'nanoid'/g" node_modules/@nuxtjs/auth/lib/schemes/oauth2.js`}).
		WithExec([]string{"yarn", "generate", "--dotenv", ".env.local"}).
		WithExec([]string{"yarn", "add", "serve"}).
		WithExposedPort(3001).
		WithExec([]string{"npx", "serve", "dist", "--debug", "--cors", "--no-port-switching", "-l", "3001"})
}

func (m *FrontendOld) Serve(ctx context.Context) *dagger.Service {
	return m.Build(ctx).AsService()
}

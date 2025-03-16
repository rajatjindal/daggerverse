package main

import (
	"context"
	"dagger/dev/internal/dagger"
	"fmt"
)

type Backend struct {
	Crud *Crud
	Src  *dagger.Directory
}

func withLocalAuth() dagger.WithContainerFunc {
	return func(c *dagger.Container) *dagger.Container {
		return c.WithEnvVariable("LOCAL_AUTH_ENABLED", "true")
	}
}

func (m *Backend) GolangVersion(ctx context.Context) string {
	version, err := dag.Toolchains().InitRequiredVersions(m.Src).Golang(ctx)
	if err != nil {
		panic(err)
	}

	return version
}

func (m *Backend) PostgresqlVersion(ctx context.Context) string {
	version, err := dag.Toolchains().InitRequiredVersions(m.Src).Postgresql(ctx)
	if err != nil {
		panic(err)
	}

	return version
}

func (m *Backend) Build(ctx context.Context) *dagger.Container {
	binary := dag.Go(dagger.GoOpts{
		Version:      m.GolangVersion(ctx),
		DisableCache: true,
		Container: dag.
			Container().
			From(fmt.Sprintf("golang:%s-alpine", m.GolangVersion(ctx))).
			WithExec([]string{"apk", "add", "git", "openssh"}).
			WithEnvVariable("GOPRIVATE", "github.com/rajatjindal/crud").
			WithExec([]string{"sh", "-c", `git config --global url.ssh://git@github.com/.insteadOf https://github.com/`}).
			WithEnvVariable("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=no ").
			WithUnixSocket("/tmp/ssh-auth-sock", m.Crud.SSHAuthSocket).
			WithEnvVariable("SSH_AUTH_SOCK", "/tmp/ssh-auth-sock")},
	).
		Build(m.Src).
		WithName(m.Crud.Name)

	return dag.Container().
		From("alpine:latest").
		WithEnvVariable("GODEBUG", "gctrace=1").
		WithFile(fmt.Sprintf("/usr/local/bin/%s", m.Crud.Name), binary).
		WithEntrypoint([]string{fmt.Sprintf("/usr/local/bin/%s", m.Crud.Name)}).
		WithExposedPort(8080).
		WithExposedPort(8081)
}

func (m *Backend) Database(ctx context.Context) *dagger.Service {
	if m.Crud.Database != nil {
		return m.Crud.Database
	}

	return dag.Container().From(fmt.Sprintf("postgres:%s", m.PostgresqlVersion(ctx))).
		WithEnvVariable("POSTGRES_DB", m.Crud.Name).
		WithEnvVariable("POSTGRES_PASSWORD", "semi-secure-password").
		WithEnvVariable("POSTGRES_USER", "postgres").
		WithEnvVariable("PGDATA", "/data/postgresql/pgdata2").
		WithFile("/docker-entrypoint-initdb.d/schema.sql", m.Src.Directory("sql").File("schema.sql")).
		WithExposedPort(5432).
		AsService()
}

func (m *Backend) Serve(ctx context.Context) *dagger.Service {
	db := m.Database(ctx)
	return m.Build(ctx).
		With(withLocalAuth()). // when running locally, disable auth
		WithServiceBinding("db.postgres.svc.cluster.local", db).
		AsService()
}

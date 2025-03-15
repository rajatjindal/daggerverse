package main

import (
	"context"
	"dagger/dev/internal/dagger"
	"fmt"
)

type Backend struct {
	Name string
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
		Version: m.GolangVersion(ctx),
		Container: dag.
			Container().
			From(fmt.Sprintf("golang:%s-alpine", m.GolangVersion(ctx)))},
	).
		Build(m.Src).
		WithName(m.Name)

	return dag.Container().
		From("alpine:latest").
		WithEnvVariable("GODEBUG", "gctrace=1").
		WithFile(fmt.Sprintf("/usr/local/bin/%s", m.Name), binary).
		WithEntrypoint([]string{fmt.Sprintf("/usr/local/bin/%s", m.Name)}).
		WithExposedPort(8080).
		WithExposedPort(8081)
}

func (m *Backend) Database(ctx context.Context) *dagger.Service {
	return dag.Container().From(fmt.Sprintf("postgres:%s", m.PostgresqlVersion(ctx))).
		WithEnvVariable("POSTGRES_DB", m.Name).
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

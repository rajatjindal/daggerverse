package main

import (
	"context"
	"dagger/dev/internal/dagger"
	"slices"
)

type Toolchain struct {
	Golang   string
	Postgres string
}

type Crud struct {
	Src       *dagger.Directory
	Toolchain Toolchain
	oldui     bool
}

func New(ctx context.Context, source *dagger.Directory) (*Crud, error) {
	crud := &Crud{
		Src: source,
	}

	entries, err := source.Entries(ctx)
	if err != nil {
		return nil, err
	}

	if slices.Contains(entries, "ui-old") {
		crud.oldui = true
	}

	return crud, nil
}

func (crud *Crud) FrontendOld() *FrontendOld {
	return &FrontendOld{
		Crud: crud,
		Src:  crud.Src.Directory("ui-old"),
	}
}

func (crud *Crud) Backend() *Backend {
	return &Backend{
		Crud: crud,
		Src:  crud.Src.Directory("backend"),
	}
}

func (crud *Crud) Frontend() *Frontend {
	return &Frontend{
		Crud: crud,
		Src:  crud.Src.Directory("ui"),
	}
}

func (crud *Crud) Prometheus() *Prometheus {
	return &Prometheus{
		Crud: crud,
	}
}

func (crud *Crud) Service(ctx context.Context) *dagger.Service {
	return crud.FrontendOld().Build(ctx).
		AsService()
}

func (crud *Crud) Serve(ctx context.Context) *dagger.Service {
	backend := crud.Backend().Serve(ctx)

	caddy := dag.Caddy().
		WithService(backend, "backend", 8080).
		WithService(backend, "backend-pprof", 8081). // pprof
		WithService(crud.Frontend().Serve(ctx), "frontend", 3000)

	if crud.oldui {
		caddy = caddy.WithService(crud.FrontendOld().Serve(ctx), "frontend-old", 3001)
	}

	return caddy.Serve()
}

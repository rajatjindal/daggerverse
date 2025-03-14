package main

import (
	"context"
	"dagger/dev/internal/dagger"
)

type Dev struct {
	Src *dagger.Directory
}

func New(ctx context.Context, source *dagger.Directory) (*Dev, error) {
	dev := &Dev{
		Src: source,
	}

	return dev, nil
}

func (dev *Dev) FrontendOld() *FrontendOld {
	return &FrontendOld{
		Dev: dev,
		Src: dev.Src.Directory("ui-old"),
	}
}

func (dev *Dev) Backend() *Backend {
	return &Backend{
		Dev: dev,
		Src: dev.Src.Directory("backend"),
	}
}

func (dev *Dev) Frontend() *Frontend {
	return &Frontend{
		Dev: dev,
		Src: dev.Src.Directory("ui"),
	}
}

func (dev *Dev) Prometheus() *Prometheus {
	return &Prometheus{
		Dev: dev,
	}
}

func (dev *Dev) Service(ctx context.Context) *dagger.Service {
	return dev.FrontendOld().Build(ctx).
		AsService()
}

func (dev *Dev) Serve(ctx context.Context) *dagger.Service {
	backend := dev.Backend().Serve(ctx)

	return dag.Caddy().
		WithService(backend, "backend", 8080).
		WithService(backend, "backend-pprof", 8081). // pprof
		WithService(dev.FrontendOld().Serve(ctx), "frontend-old", 3001).
		WithService(dev.Frontend().Serve(ctx), "frontend", 3000).
		Serve()
}

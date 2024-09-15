package main

import (
	"context"
	"fmt"

	"dagger/caddy/internal/dagger"
)

type Caddy struct {
	Services []*ServiceConfig
}

type ServiceConfig struct {
	UpstreamName string
	UpstreamPort int32
	UpstreamSvc  *dagger.Service
}

func New() *Caddy {
	return &Caddy{
		Services: []*ServiceConfig{},
	}
}

func (c *Caddy) WithService(ctx context.Context, upstreamService *dagger.Service, upstreamName string, upstreamPort int32) *Caddy {
	c.Services = append(c.Services, &ServiceConfig{
		UpstreamName: upstreamName,
		UpstreamPort: upstreamPort,
		UpstreamSvc:  upstreamService,
	})

	return c
}

func (c *Caddy) GetCaddyFile(ctx context.Context) string {
	caddyFile := ""
	for _, svc := range c.Services {
		caddyFile += fmt.Sprintf(`
:%d {
		reverse_proxy %s:%d
}

`, svc.UpstreamPort, svc.UpstreamName, svc.UpstreamPort)
	}

	return caddyFile
}

func (c *Caddy) Container(ctx context.Context) *dagger.Container {
	ctr := dag.Container().From("caddy:2.8.4").
		WithNewFile("/opt/caddy/caddyfile", c.GetCaddyFile(ctx))

	for _, svc := range c.Services {
		ctr = ctr.WithServiceBinding(svc.UpstreamName, svc.UpstreamSvc)
	}

	ctr.
		WithExec([]string{"caddy", "run", "--config", "/opt/caddy/caddyfile"}).
		WithExposedPort(80)

	return ctr
}

func (c *Caddy) Serve(ctx context.Context) *dagger.Service {
	return c.Container(ctx).AsService()
}

package main

import (
	"context"
	"fmt"

	"dagger/caddy/internal/dagger"
)

type Caddy struct {
	Services []*Service
}

type Service struct {
	LBDomainName string
	UpstreamName string
	UpstreamPort int32
	UpstreamSvc  *dagger.Service
}

func New() *Caddy {
	return &Caddy{
		Services: []*Service{},
	}
}

func (c *Caddy) WithService(ctx context.Context, upstreamService *dagger.Service, lbDomainName, upstreamName string, upstreamPort int32) *Caddy {
	c.Services = append(c.Services, &Service{
		LBDomainName: lbDomainName,
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
http://%s {
		reverse_proxy %s:%d
}

`, svc.LBDomainName, svc.UpstreamName, svc.UpstreamPort)
	}

	return caddyFile
}

func (c *Caddy) Service(ctx context.Context) *dagger.Service {
	return dag.Container().From("caddy:2.8.4").
		WithNewFile("/opt/caddy/caddyfile", c.GetCaddyFile(ctx)).
		WithExec([]string{"caddy", "run", "--config", "/opt/caddy/caddyfile"}).
		AsService()
}

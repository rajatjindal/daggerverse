package main

import (
	"context"
	"dagger/dev/internal/dagger"
)

type Prometheus struct {
	Crud *Crud
}

func (m *Prometheus) Build(ctx context.Context) *dagger.Container {
	return dag.Container().
		From("prom/prometheus").
		WithNewFile("/etc/prometheus/prometheus.yml", `---
scrape_configs:
- job_name: 'backend'
  scrape_interval: 15s
  metrics_path: '/metrics'
  static_configs:
    - targets: ['backend:8081']
`).WithExposedPort(9090)
}

func (m *Prometheus) Serve(ctx context.Context, backend *dagger.Service) *dagger.Service {
	return m.Build(ctx).
		WithServiceBinding("backend", backend).
		AsService(dagger.ContainerAsServiceOpts{
			UseEntrypoint: true,
			Args:          ([]string{"--config.file=/etc/prometheus/prometheus.yml"}),
		})
}

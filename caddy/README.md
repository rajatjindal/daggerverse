Usage:

### Install

```
$ dagger use github.com/rajatjindal/daggerverse/caddy@main
```
### Use
```go
	return dag.Proxy().
		WithService(dev.Backend().Serve(ctx), "backend", 8080).
		WithService(dev.Frontend().Serve(ctx), "frontend", 3000).
		Serve()
```
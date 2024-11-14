Usage:

### Use
Create a file `.toolchains` in your Spin App directory

e.g.

```
spin=3.0.0
golang=1.23.2
tinygo=0.34.0
```

and then run `dagger call -m github.com/rajatjindal/daggerverse/wasi@main --source=. terminal`
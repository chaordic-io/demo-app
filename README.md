# Super simple Go web server with full observability
Bare minimum app to demonstrate full observability with the Grafana-stack (Prometheus, Tempo & Loki) from an app perspective running on Nomad.
`main.go`, `demo-app.hcl` and `Dockerfile` can all be optimized, use more precise options and more DRY, thus should only serve as examples spelling out the minimum of what needs to be done.

### Other
Build on Apple Silicon:
`docker buildx build --platform linux/amd64 -t demo-app .`
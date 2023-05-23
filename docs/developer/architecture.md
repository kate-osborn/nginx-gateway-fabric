# Architecture and Design

## Project Structure

* NGINX Kubernetes Gateway is written in Go and uses the open source NGINX software as the data plane.
* The project follows a standard Go project layout
    * The main code is found at `cmd/gateway/`
    * The internal code is found at `internal/`
    * Build files for Docker are found under `build/`
    * Deployment yaml files are found at `deploy/`
    * External APIs, clients, and SDKs can be found under `pkg/`
* We use [Go Modules](https://github.com/golang/go/wiki/Modules) for managing dependencies.
* We use [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/) for our BDD style unit
  tests.

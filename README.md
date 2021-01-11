# Cobweb

A simple Envoy Proxy Control Plane 


* Dockerhub Page
    * https://hub.docker.com/repository/docker/lfmunoz4/cobweb


# Usage

```bash
# docker
docker run -it --rm lfmunoz4/cobweb:0.1.0

# binary (overwrite configs using ./conf.json)
./main 
```

# Development

See Makefile

```bash
make test
make run
make build
```

# Technologies

* Go
    * go version go1.15.5 linux/amd64
* Vuejs
* Envoy Proxy
* Git Actions
    * https://github.com/actions/setup-go




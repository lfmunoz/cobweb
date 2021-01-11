# ________________________________________________________________________________
# DEVELOPMENT
# ________________________________________________________________________________
.PHONY: run test unit

run: cmd/main/main.go
	go run $<

example: cmd/example/example.go
	go run $<

test:
	go test -v ./...

unit:
	go test ./...

# ________________________________________________________________________________
# PRODUCTION
# ________________________________________________________________________________
.PHONY: clean docker
COBWEB_VERSION=0.1.0

main: cmd/main/main.go
	# go build cmd/main/main.go
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build cmd/main/main.go

# Build application 
build: main

# Build docker image
build-docker: build
	@echo "Deploy to dockerhub"
	cp main docker/main
	cp web -r docker/web
	cd docker; docker build -t lfmunoz4/cobweb:${COBWEB_VERSION} .

# Upload to dockerhub
# docker pull golang:1.15.5
# docker run -it --rm golang:1.15.5 /bin/bash
deploy: build-docker
	docker push lfmunoz4/cobweb:${COBWEB_VERSION}

clean:
	-docker image rm lfmunoz4/cobweb:${COBWEB_VERSION}
	-rm main
	-rm docker/main
	-rm -rf docker/web

docker:
	#docker run -it --rm lfmunoz4/cobweb:${COBWEB_VERSION} 
	docker run -it --rm --network=host lfmunoz4/cobweb:${COBWEB_VERSION} 

# ________________________________________________________________________________
# INFO
# ________________________________________________________________________________
.PHONY: info

info:
	go version


# ________________________________________________________________________________
# Start Ubuntu on a docker instance 
# ________________________________________________________________________________
start: 
	-docker start envoy
	# <hostPort>:<containerPort>
	#  3080 on loclahost will go to 80 inside container
	-docker run -d -p 3080:80 -p 9901:9901 --name envoy --privileged \
		-v $(PWD)/:/root/tools \
		lfmunoz4/bertha:2.0.0

bash:
	docker exec -it envoy /bin/bash

stop: 
	-docker stop envo

e1:
	envoy -c dynamic.yaml

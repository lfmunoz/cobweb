# ________________________________________________________________________________
# DEVELOPMENT
# ________________________________________________________________________________
.PHONY: run

run: cmd/server/server.go
	go run $<

example: cmd/example/example.go
	go run $<

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

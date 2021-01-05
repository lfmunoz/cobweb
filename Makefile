
# ________________________________________________________________________________
# INFO
# ________________________________________________________________________________
info:
	go version


.PHONY: run

run: cmd/server/main.go
	go run $<

example: cmd/example/example.go
	go run $<

HOST=localhost
PORT=3080

# --resolve <host:port:address> Force resolve of HOST:PORT to ADDRESS
# --resolve  localhost:3080:127.0.0.1 
api:
	curl  -v -H "Host: 0.0.0.0" \
   		--cacert certs/tls-ca.crt https://localhost:3080/


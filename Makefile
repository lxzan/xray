PWD=$(shell pwd)

tidy:
	cd ${PWD}/contrib/adapter/gws && go mod tidy
	cd ${PWD}/contrib/adapter/http && go mod tidy
	cd ${PWD}/contrib/adapter/fasthttp && go mod tidy
	cd ${PWD}/contrib/codec/jsoniter && go mod tidy
	cd ${PWD}/contrib/codec/wwwform && go mod tidy
	cd ${PWD}/contrib/doc/swagger && go mod tidy
	cd ${PWD}/contrib/log/zerolog && go mod tidy
	cd ${PWD}/examples/http_server && go mod tidy
	cd ${PWD}/examples/http3_server && go mod tidy
	cd ${PWD}/examples/fasthttp_server && go mod tidy
	cd ${PWD}/examples/gws_server && go mod tidy
	go mod tidy

test:
	cd ${PWD}/contrib/adapter/gws && go test --count=1 ./...
	cd ${PWD}/contrib/adapter/http && go test --count=1 ./...
	cd ${PWD}/contrib/codec/jsoniter && go test --count=1 ./...
	cd ${PWD}/contrib/codec/wwwform && go test --count=1 ./...
	cd ${PWD}/contrib/log/zerolog && go test --count=1 ./...
	cd ${PWD}/contrib/doc/swagger && go test --count=1 ./...
	go test --count=1 ./...

cover:
	go test -coverprofile=./bin/cover.out --cover ./...

bench:
	go test -benchmem -run=^$$ -bench . github.com/lxzan/xray

build-linux:
	GOOS=linux go build -o ./bin/gws.linux ./examples/gws_server/main.go
	GOOS=linux go build -o ./bin/http.linux github.com/lxzan/xray/examples/http_server
	GOOS=linux go build -o ./bin/http3.linux ./examples/http3_server/main.go
	GOOS=linux go build -o ./bin/fasthttp.linux ./examples/fasthttp_server/main.go

clean:
	rm ./bin/*

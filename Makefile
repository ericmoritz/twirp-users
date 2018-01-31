all: twirp test build

twirp:
	protoc --proto_path=$$GOPATH/src:. --twirp_out=. --go_out=. ./rpc/users/service.proto

test:
	go test ./tests

build:
	go build

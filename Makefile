all: twirp test build

twirp:
	protoc --proto_path=$$GOPATH/src:. --twirp_out=. --go_out=. ./rpc/users/service.proto

test:
	govendor test +local

build:
	go build

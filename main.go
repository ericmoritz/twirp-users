package main

import (
	"github.com/ericmoritz/twirp-users/internal/usersservice"
	pb "github.com/ericmoritz/twirp-users/rpc/users"
	"net/http"
	"os"
	"fmt"
)

func main() {
	server, err := usersservice.New("./.usersservice.db")
	if err != nil {
		panic(err)
	}



	var bind = ":8080"
	if port := os.Getenv("PORT"); port != "" {
		bind = ":"+port
	}


	handler := pb.NewUsersServer(server, nil)
	fmt.Printf("Listening on %s\n", bind)
	err = http.ListenAndServe(bind, handler)
	if err != nil {
		panic(err)
	}
}

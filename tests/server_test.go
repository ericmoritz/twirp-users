package usersservice_test

import (
	"context"
	"testing"
	. "github.com/franela/goblin"
	pb "github.com/ericmoritz/twirp-users/rpc/users"
	"github.com/twitchtv/twirp"
	"github.com/ericmoritz/twirp-users/internal/usersservice"
	"os"
)

// Test tests the server
func Test(t *testing.T) {
	g := Goblin(t)

	g.Describe("Users API", func() {
		var service pb.Users
		testDbPath := "/tmp/usersservice.db"

		g.Before(func() {
			// Delete the db if it exists
			if _, err := os.Stat(testDbPath); err == nil {
				if err := os.RemoveAll(testDbPath); err != nil {
					panic(err)
				}
			}


			if s, err := usersservice.New(testDbPath); err == nil {
				service = s
			} else {
				panic(err)
			}
		})

		g.It("Happy Case", func() {
			// Test registration
			resp, err := service.Register(context.Background(), &pb.RegisterReq{Username: "eric", Password: "Shhh"})
			g.Assert(err).Equal(nil)
			g.Assert(resp.User.Username).Equal("eric")

			// Test login
			loginResp, err := service.Login(context.Background(), &pb.LoginReq{Username: "eric", Password: "Shhh"})
			g.Assert(err).Equal(nil)

			// Test User request
			userResp, err := service.User(
				context.Background(),
				&pb.UserReq{
					Session: loginResp.Session,
					Username: "eric",
				},
			)
			g.Assert(err).Equal(nil)
			g.Assert(userResp.User.Username).Equal("eric")

			// Test CurrentUser request
			currentUserResp, err := service.CurrentUser(
				context.Background(),
				&pb.CurrentUserReq{
					Session: loginResp.Session,
				},
			)
			g.Assert(err).Equal(nil)
			g.Assert(currentUserResp.User.Username).Equal("eric")
		})

		g.It("Should fail to register if the username is blank", func() {
			_, err := service.Register(context.Background(), &pb.RegisterReq{Username: "", Password: "Shhh"})
			g.Assert(err).Equal(twirp.RequiredArgumentError("RegisterReq.username"))
		})

		g.It("Should fail to register if the password is blank", func() {
			_, err := service.Register(context.Background(), &pb.RegisterReq{Username: "eric", Password: ""})
			g.Assert(err).Equal(twirp.RequiredArgumentError("RegisterReq.password"))
		})

		// TODO the rest of the owl.
	})
}

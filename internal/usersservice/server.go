package usersservice

import (
	"context"
	"github.com/twitchtv/twirp"
	"github.com/syndtr/goleveldb/leveldb"
	pb "github.com/ericmoritz/twirp-users/rpc/users"
	"crypto/sha256"
	"github.com/golang/protobuf/proto"
	"bytes"
	"github.com/satori/go.uuid"
)

// Create a new userService
func New(dbPath string) (*userService, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}

	return &userService{DB: db}, nil
}

type userService struct {
	DB *leveldb.DB
}

// Register registers a user
func (us *userService) Register(c context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
	// Validate the username
	if req.Username == "" {
		return nil, twirp.RequiredArgumentError("RegisterReq.username")
	}

	if req.Password == "" {
		return nil, twirp.RequiredArgumentError("RegisterReq.password")
	}

	////
	// Create the User
	////
	user := &pb.PrivateUser{
		Username: req.Username,
		PasswordSha256: hashPassword(req.Password),
	}

	////
	// Store the user
	////
	if err := putUser(us.DB, user); err != nil {
		return nil, err
	}

	// Return the response
	return &pb.RegisterResp{
		User: &pb.User{
			Username: user.Username,
		},
	}, nil
}

func (us *userService) Login(c context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	// Find the user
	user, err := getUser(us.DB, req.Username)
	if err != nil {
		return nil, err
	}

	// If the username in blank, the user does not exist
	if user.Username == "" {
		return nil, twirp.NewError(twirp.PermissionDenied, "bad username")
	}

	// Check the passwords
	if bytes.Compare(user.PasswordSha256, hashPassword(req.Password)) != 0 {
		return nil, twirp.NewError(twirp.PermissionDenied, "bad password")
	}

	// Login successful, create a session token
	session := &pb.Session{
		Token: uuid.NewV4().String(),
		Username: user.Username,
	}
	// Store the session
	if err := putSession(us.DB, session); err != nil {
		return nil, err
	}

	return &pb.LoginResp{
		Session: session,
	}, nil
}

func (us *userService) User(c context.Context, req *pb.UserReq) (*pb.UserResp, error) {
	user, err := getUser(us.DB, req.Username)
	if err != nil {
		return nil, err
	}

	return &pb.UserResp{
		User: &pb.User{
			Username: user.Username,
		},
	}, nil
}

func (us *userService) CurrentUser(c context.Context, req *pb.CurrentUserReq) (*pb.CurrentUserResp, error) {
	session, err := validateSession(us.DB, req.Session)
	if err != nil {
		return nil, err
	}
	user, err := getUser(us.DB, session.Username)
	if err != nil {
		return nil, err
	}
	return &pb.CurrentUserResp{
		User: &pb.User{
			Username: user.Username,
		},
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// Internal
///////////////////////////////////////////////////////////////////////////////
func validateSession(db *leveldb.DB, session *pb.Session) (*pb.Session, error) {
	_, err := getSession(db, session.Token)
	if err == leveldb.ErrNotFound {
		return nil, twirp.NewError(twirp.PermissionDenied, "invalid session token")
	} else if err != nil {
		return nil, err
	}
	return session, nil
}

func hashPassword(password string) []byte {
	// Sha the password
	h := sha256.New()
	h.Write([]byte(password))
	return h.Sum(nil)
}


func getUser(db *leveldb.DB, username string) (*pb.PrivateUser, error) {
	bytes, err := db.Get(userKey(username), nil)
	if err == leveldb.ErrNotFound {
		return nil, twirp.NewError(twirp.NotFound, username + " not found")
	} else if err != nil {
		return nil, err
	}

	user := &pb.PrivateUser{}
	if err := proto.Unmarshal(bytes, user); err != nil {
		return nil, err
	}
	return user, nil
}


func putUser(db *leveldb.DB, user *pb.PrivateUser) error {
	exists, err := db.Has(userKey(user.Username), nil)
	if err != nil {
		return err
	}
	if exists == true {
		return twirp.NewError(twirp.AlreadyExists, "Username: " + user.Username + " already exists")
	}

	// Store the user into the db
	bytes, err := proto.Marshal(user)
	if err != nil {
		return err
	}

	return db.Put(userKey(user.Username), bytes, nil)
}


func putSession(db *leveldb.DB, session *pb.Session) error {
	bytes, err := proto.Marshal(session)
	if err != nil {
		return err
	}

	return db.Put(sessionKey(session.Token), bytes, nil)
}


func getSession(db *leveldb.DB, token string) (*pb.Session, error) {
	bytes, err := db.Get(sessionKey(token), nil)
	if err != nil {
		return nil, err
	}

	session := &pb.Session{}
	if err := proto.Unmarshal(bytes, session); err != nil {
		return nil, err
	}
	return session, nil
}

func userKey(username string) []byte {
	return []byte("users/" + username)
}

func sessionKey(token string) []byte {
	return []byte("sessions/" + token)
}

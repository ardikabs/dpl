package types

import (
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitSecret struct {
	Username string
	Password string
}

func (o GitSecret) GetAuthMethod() transport.AuthMethod {
	return &http.BasicAuth{Username: o.Username, Password: o.Password}
}

func (o GitSecret) IsZero() bool {
	return o.Username == "" || o.Password == ""
}

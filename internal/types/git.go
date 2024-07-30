package types

import "fmt"

type GitSecret struct {
	Username string
	Password string
}

func (o GitSecret) Raw() string {
	return fmt.Sprintf("%s:%s", o.Username, o.Password)
}

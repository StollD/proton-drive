package drive

import (
	"context"
	"github.com/henrybear327/go-proton-api"
)

type InitHandler func(ctx context.Context) error

type Session struct {
	application *Application

	user *User
}

func NewSession(application *Application) *Session {
	self := &Session{application: application}

	self.user = &User{client: self.Client(), tokens: self.Tokens()}

	return self
}

func (self *Session) Client() *proton.Client {
	return self.application.Client()
}

func (self *Session) Tokens() *Tokens {
	return self.application.Tokens()
}

func (self *Session) User() *User {
	return self.user
}

func (self *Session) Init(ctx context.Context) error {
	err := self.user.Init(ctx)
	if err != nil {
		return err
	}

	return nil
}

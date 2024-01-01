package drive

import (
	"context"
	"github.com/henrybear327/go-proton-api"
)

type InitHandler func(ctx context.Context) error

type Session struct {
	application *Application
}

func NewSession(application *Application) *Session {
	self := &Session{application: application}

	return self
}

func (self *Session) Client() *proton.Client {
	return self.application.Client()
}

func (self *Session) Tokens() *Tokens {
	return self.application.Tokens()
}

func (self *Session) Init(ctx context.Context) error {
	return nil
}

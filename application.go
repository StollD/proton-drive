package drive

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/henrybear327/go-proton-api"
)

var (
	ErrUsernamePasswordMissing = errors.New("username/password missing")
	ErrTwoFactorTokenMissing   = errors.New("2fa token missing")
	ErrMailboxPasswordMissing  = errors.New("mailbox password missing")
)

type TokensUpdatedHandler func(*Tokens)
type TokensExpiredHandler func()

type Application struct {
	manager *proton.Manager
	client  *proton.Client

	tokens *Tokens

	onTokensUpdated []TokensUpdatedHandler
	onTokensExpired []TokensExpiredHandler
}

func NewApplication(version string) *Application {
	manager := proton.New(proton.WithAppVersion(version))

	return &Application{
		manager:         manager,
		onTokensUpdated: []TokensUpdatedHandler{},
		onTokensExpired: []TokensExpiredHandler{},
	}
}

func (self *Application) LoginWithCredentials(ctx context.Context, credentials Credentials) error {
	username := credentials.Username
	password := credentials.Password

	bytePass := []byte(password)

	if username == "" || password == "" {
		return ErrUsernamePasswordMissing
	}

	client, auth, err := self.manager.NewClientWithLogin(ctx, username, bytePass)
	if err != nil {
		return err
	}

	self.client = client

	if auth.TwoFA.Enabled&proton.HasTOTP != 0 {
		if credentials.TwoFA == "" {
			return ErrTwoFactorTokenMissing
		}

		err := client.Auth2FA(ctx, proton.Auth2FAReq{
			TwoFactorCode: credentials.TwoFA,
		})

		if err != nil {
			return err
		}
	}

	keyPass := bytePass

	if auth.PasswordMode == proton.TwoPasswordMode {
		if credentials.MailboxPassword == "" {
			return ErrMailboxPasswordMissing
		}

		keyPass = []byte(credentials.MailboxPassword)
	}

	salts, err := client.GetSalts(ctx)
	if err != nil {
		return err
	}

	user, err := client.GetUser(ctx)
	if err != nil {
		return err
	}

	saltedBytes, err := salts.SaltForKey(keyPass, user.Keys.Primary().ID)
	if err != nil {
		return err
	}

	self.tokens = &Tokens{
		UID:           auth.UID,
		AccessToken:   auth.AccessToken,
		RefreshToken:  auth.RefreshToken,
		SaltedKeyPass: base64.StdEncoding.EncodeToString(saltedBytes),
	}

	self.client.AddAuthHandler(func(auth proton.Auth) {
		self.callOnTokensUpdated(auth)
	})

	self.client.AddDeauthHandler(func() {
		self.callOnTokensExpired()
	})

	return nil
}

func (self *Application) LoginWithTokens(tokens *Tokens) {
	self.tokens = tokens
	self.client = self.manager.NewClient(tokens.UID, tokens.AccessToken, tokens.RefreshToken)

	self.client.AddAuthHandler(func(auth proton.Auth) {
		self.callOnTokensUpdated(auth)
	})

	self.client.AddDeauthHandler(func() {
		self.callOnTokensExpired()
	})
}

func (self *Application) Manager() *proton.Manager {
	return self.manager
}

func (self *Application) Client() *proton.Client {
	return self.client
}

func (self *Application) Tokens() *Tokens {
	return self.tokens
}

func (self *Application) OnTokensUpdated(handler TokensUpdatedHandler) {
	self.onTokensUpdated = append(self.onTokensUpdated, handler)
}

func (self *Application) OnTokensExpired(handler TokensExpiredHandler) {
	self.onTokensExpired = append(self.onTokensExpired, handler)
}

func (self *Application) callOnTokensUpdated(auth proton.Auth) {
	tokens := &Tokens{
		UID:           auth.UID,
		AccessToken:   auth.AccessToken,
		RefreshToken:  auth.RefreshToken,
		SaltedKeyPass: self.tokens.SaltedKeyPass,
	}

	self.tokens = tokens

	for _, handler := range self.onTokensUpdated {
		handler(tokens)
	}
}

func (self *Application) callOnTokensExpired() {
	for _, handler := range self.onTokensExpired {
		handler()
	}
}

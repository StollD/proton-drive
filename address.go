package drive

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/henrybear327/go-proton-api"
)

type Address struct {
	address proton.Address
	keyring *crypto.KeyRing
}

func (self *Address) ID() string {
	return self.address.ID
}

func (self *Address) Email() string {
	return self.address.Email
}

func (self *Address) Keyring() *crypto.KeyRing {
	return self.keyring
}

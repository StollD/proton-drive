package drive

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/henrybear327/go-proton-api"
)

type Share struct {
	share   proton.Share
	address *Address
	keyring *crypto.KeyRing
}

func (self *Share) ID() string {
	return self.share.ShareID
}

func (self *Share) LinkID() string {
	return self.share.LinkID
}

func (self *Share) Address() *Address {
	return self.address
}

func (self *Share) Keyring() *crypto.KeyRing {
	return self.keyring
}

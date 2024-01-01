package drive

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/henrybear327/go-proton-api"
)

var (
	ErrKeyringUnlockFailed = errors.New("failed to unlock keyring")
)

type User struct {
	//
	// PARAMETERS
	//

	client *proton.Client
	tokens *Tokens

	//
	// INTERNAL STATE
	//

	user    proton.User
	keyring *crypto.KeyRing

	addresses      []*Address
	addressByID    map[string]*Address
	addressByEmail map[string]*Address
}

func (self *User) Init(ctx context.Context) error {
	user, err := self.client.GetUser(ctx)
	if err != nil {
		return err
	}

	addresses, err := self.client.GetAddresses(ctx)
	if err != nil {
		return err
	}

	pass, err := base64.StdEncoding.DecodeString(self.tokens.SaltedKeyPass)
	if err != nil {
		return err
	}

	keyring, addrKeyrings, err := proton.Unlock(user, addresses, pass, nil)
	if err != nil {
		return err
	}

	if keyring.CountDecryptionEntities() == 0 {
		return ErrKeyringUnlockFailed
	}

	self.user = user
	self.keyring = keyring

	self.addresses = []*Address{}
	self.addressByID = map[string]*Address{}
	self.addressByEmail = map[string]*Address{}

	for _, addr := range addresses {
		address := &Address{
			address: addr,
			keyring: addrKeyrings[addr.ID],
		}

		self.addresses = append(self.addresses, address)
		self.addressByID[address.ID()] = address
		self.addressByEmail[address.Email()] = address
	}

	return nil
}

func (self *User) Keyring() *crypto.KeyRing {
	return self.keyring
}

func (self *User) Addresses() []*Address {
	return self.addresses
}

func (self *User) AddressFromID(id string) *Address {
	if val, ok := self.addressByID[id]; ok {
		return val
	}

	return nil
}

func (self *User) AddressFromEmail(email string) *Address {
	if val, ok := self.addressByEmail[email]; ok {
		return val
	}

	return nil
}

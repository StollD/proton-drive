package drive

import (
	pathlib "path"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/henrybear327/go-proton-api"
)

type Link struct {
	link proton.Link

	name  string
	share *Share
	revID string

	signAddress     *Address
	nameSignAddress *Address

	parent   *Link
	children mapset.Set[*Link]

	attrs      *Attributes
	keyring    *crypto.KeyRing
	sessionKey *crypto.SessionKey

	hashKey []byte
}

func (self *Link) ID() string {
	return self.link.LinkID
}

func (self *Link) Name() string {
	if self.IsRoot() {
		return "/"
	}

	return self.name
}

func (self *Link) Path() string {
	if self.IsRoot() {
		return self.Name()
	}

	return pathlib.Join(self.parent.Path(), self.Name())
}

func (self *Link) Share() *Share {
	return self.share
}

func (self *Link) RevisionID() string {
	return self.revID
}

func (self *Link) SignatureAddress() *Address {
	return self.signAddress
}

func (self *Link) NameSignatureEmail() *Address {
	return self.nameSignAddress
}

func (self *Link) Parent() *Link {
	return self.parent
}

func (self *Link) Children() mapset.Set[*Link] {
	return self.children
}

func (self *Link) IsFile() bool {
	return self.link.Type == proton.LinkTypeFile
}

func (self *Link) IsDir() bool {
	return self.link.Type == proton.LinkTypeFolder
}

func (self *Link) IsRoot() bool {
	return self.parent == nil
}

func (self *Link) Size() int64 {
	if self.attrs == nil {
		return self.link.Size
	}

	return self.attrs.Size
}

func (self *Link) Hash() string {
	return self.link.Hash
}

func (self *Link) ContentHash() string {
	if self.attrs == nil {
		return ""
	}

	return self.attrs.Hash
}

func (self *Link) MIMEType() string {
	if self.IsDir() {
		return "inode/directory"
	}

	if self.attrs == nil {
		return ""
	}

	return self.attrs.MIMEType
}

func (self *Link) BlockSizes() []int64 {
	return self.attrs.BlockSizes
}

func (self *Link) CreationTime() time.Time {
	return time.Unix(self.link.CreateTime, 0)
}

func (self *Link) ModificationTime() time.Time {
	if self.attrs == nil {
		return time.Unix(self.link.ModifyTime, 0)
	}

	return self.attrs.ModifyTime
}

func (self *Link) Keyring() *crypto.KeyRing {
	return self.keyring
}

func (self *Link) SessionKey() *crypto.SessionKey {
	return self.sessionKey
}

func (self *Link) HashKey() []byte {
	return self.hashKey
}

func (self *Link) NodePassphrase() string {
	return self.link.NodePassphrase
}

func (self *Link) NodePassphraseSignature() string {
	return self.link.NodePassphraseSignature
}

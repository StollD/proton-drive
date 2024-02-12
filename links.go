package drive

import (
	"context"
	"errors"
	pathlib "path"

	"github.com/barweiss/go-tuple"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/henrybear327/go-proton-api"
	"github.com/relvacode/iso8601"
	"golang.org/x/time/rate"
)

var (
	ErrMainVolumeNotFound             = errors.New("main volume not found")
	ErrShareAddressNotFound           = errors.New("share address not found")
	ErrLinkSignatureEmailNotFound     = errors.New("signature email not found")
	ErrLinkNameSignatureEmailNotFound = errors.New("name signature email not found")
)

type Links struct {
	//
	// PARAMETERS
	//

	client *proton.Client
	user   *User

	//
	// INTERNAL STATE
	//

	volume *Volume
	share  *Share
	root   *Link

	linkByID   map[string]*Link
	linkByPath map[string]*Link

	limiter *rate.Limiter
}

func (self *Links) Init(ctx context.Context) error {
	err := self.getVolume(ctx)
	if err != nil {
		return err
	}

	err = self.getShare(ctx)
	if err != nil {
		return err
	}

	err = self.getRoot(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (self *Links) getVolume(ctx context.Context) error {
	volumes, err := self.client.ListVolumes(ctx)
	if err != nil {
		return err
	}

	for _, volume := range volumes {
		if volume.State != proton.VolumeStateActive {
			continue
		}

		self.volume = &Volume{volume: volume}
		return nil
	}

	return ErrMainVolumeNotFound
}

func (self *Links) getShare(ctx context.Context) error {
	share, err := self.client.GetShare(ctx, self.volume.ShareID())
	if err != nil {
		return err
	}

	address := self.user.AddressFromID(share.AddressID)
	if address == nil {
		return ErrShareAddressNotFound
	}

	keyring, err := share.GetKeyRing(address.Keyring())
	if err != nil {
		return err
	}

	self.share = &Share{
		share:   share,
		address: address,
		keyring: keyring,
	}

	return nil
}

func (self *Links) getRoot(ctx context.Context) error {
	self.limiter = rate.NewLimiter(8, 1)

	rootLink, err := self.client.GetLink(ctx, self.share.ID(), self.share.LinkID())
	if err != nil {
		return err
	}

	root, err := self.getLinksRecursive(ctx, rootLink, nil)
	if err != nil {
		return err
	}

	self.root = root
	self.getLinkMaps()

	return nil
}

func (self *Links) getLinksRecursive(ctx context.Context, link proton.Link, parent *Link) (*Link, error) {
	out, err := self.getLink(link, parent)
	if err != nil {
		return nil, err
	}

	if out.IsFile() {
		return out, nil
	}

	err = self.limiter.Wait(ctx)
	if err != nil {
		return nil, err
	}

	childLinks, err := self.client.ListChildren(ctx, self.share.ID(), out.ID(), true)
	if err != nil {
		return nil, err
	}

	count := 0
	channel := make(chan tuple.T2[*Link, error])

	for i := range childLinks {
		childLink := childLinks[i]

		if childLink.State != proton.LinkStateActive {
			continue
		}

		count++

		go func() {
			channel <- tuple.New2(self.getLinksRecursive(ctx, childLink, out))
		}()
	}

	for i := 0; i < count; i++ {
		ret := <-channel
		if ret.V2 != nil {
			return nil, ret.V2
		}

		out.children.Add(ret.V1)
	}

	return out, nil
}

func (self *Links) getLink(link proton.Link, parent *Link) (*Link, error) {
	signAddress := self.user.AddressFromEmail(link.SignatureEmail)
	if signAddress == nil {
		return nil, ErrLinkSignatureEmailNotFound
	}

	nameSignAddress := self.user.AddressFromEmail(link.NameSignatureEmail)
	if nameSignAddress == nil {
		return nil, ErrLinkNameSignatureEmailNotFound
	}

	parentKR := self.share.Keyring()
	if parent != nil {
		parentKR = parent.Keyring()
	}

	keyring, err := link.GetKeyRing(parentKR, signAddress.Keyring())
	if err != nil {
		return nil, err
	}

	name, err := link.GetName(parentKR, nameSignAddress.Keyring())
	if err != nil {
		return nil, err
	}

	xAttrs, err := link.GetDecXAttrString(signAddress.Keyring(), keyring)
	if err != nil {
		return nil, err
	}

	out := &Link{
		link: link,

		name:  name,
		share: self.share,

		signAddress:     signAddress,
		nameSignAddress: nameSignAddress,

		parent:   parent,
		children: mapset.NewSet[*Link](),

		attrs:   nil,
		keyring: keyring,
	}

	if out.IsFile() {
		out.revID = link.FileProperties.ActiveRevision.ID

		sessionKey, err := link.GetSessionKey(keyring)
		if err != nil {
			return nil, err
		}

		out.sessionKey = sessionKey
	} else {
		hashKey, err := link.GetHashKey(keyring, signAddress.Keyring())
		if err != nil {
			return nil, err
		}

		out.hashKey = hashKey
	}

	if xAttrs != nil {
		modTime, err := iso8601.ParseString(xAttrs.ModificationTime)
		if err != nil {
			return nil, err
		}

		out.attrs = &Attributes{
			Size:       xAttrs.Size,
			Hash:       xAttrs.Digests["SHA1"],
			MIMEType:   link.MIMEType,
			ModifyTime: modTime,
			BlockSizes: xAttrs.BlockSizes,
		}
	}

	return out, nil
}

func (self *Links) getLinkMaps() {
	self.linkByID = map[string]*Link{}
	self.linkByPath = map[string]*Link{}

	self.getLinkMapsRecursive(self.root)
}

func (self *Links) getLinkMapsRecursive(link *Link) {
	self.linkByID[link.ID()] = link
	self.linkByPath[link.Path()] = link

	for child := range link.Children().Iter() {
		self.getLinkMapsRecursive(child)
	}
}

func (self *Links) Volume() *Volume {
	return self.volume
}

func (self *Links) Share() *Share {
	return self.share
}

func (self *Links) Root() *Link {
	return self.root
}

func (self *Links) LinkFromID(linkID string) *Link {
	if val, ok := self.linkByID[linkID]; ok {
		return val
	}

	return nil
}

func (self *Links) LinkFromPath(path string) *Link {
	path = pathlib.Clean(path)

	if val, ok := self.linkByPath[path]; ok {
		return val
	}

	return nil
}

func (self *Links) OnEvent(event proton.LinkEvent) error {
	old := self.LinkFromID(event.Link.LinkID)

	if event.Link.State == proton.LinkStateActive {
		if old == nil {
			return self.onCreate(event)
		} else {
			return self.onUpdate(event)
		}
	} else {
		if old != nil {
			self.onDelete(event)
		}
	}

	return nil
}

func (self *Links) onCreate(event proton.LinkEvent) error {
	if event.Link.State != proton.LinkStateActive {
		return nil
	}

	parent := self.LinkFromID(event.Link.ParentLinkID)

	link, err := self.getLink(event.Link, parent)
	if err != nil {
		return err
	}

	parent.children.Add(link)

	self.linkByID[link.ID()] = link
	self.linkByPath[link.Path()] = link

	return nil
}

func (self *Links) onUpdate(event proton.LinkEvent) error {
	old := self.LinkFromID(event.Link.LinkID)

	oldParent := old.Parent()
	newParent := self.LinkFromID(event.Link.ParentLinkID)

	link, err := self.getLink(event.Link, newParent)
	if err != nil {
		return err
	}

	link.children = old.children
	*old = *link

	oldParent.children.Remove(old)
	newParent.children.Add(old)

	self.getLinkMaps()
	return nil
}

func (self *Links) onDelete(event proton.LinkEvent) {
	old := self.LinkFromID(event.Link.LinkID)

	if old == nil {
		return
	}

	delete(self.linkByID, old.ID())
	delete(self.linkByPath, old.Path())

	old.Parent().children.Remove(old)
}

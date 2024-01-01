package drive

import (
	"context"
	"errors"
	"github.com/henrybear327/go-proton-api"
)

var (
	ErrInvalidLink     = errors.New("invalid link")
	ErrInvalidLinkType = errors.New("invalid link type, expected file")
)

type FileSystem struct {
	//
	// PARAMETERS
	//

	client *proton.Client
	user   *User
	links  *Links
	events *EventLoop
}

func (self *FileSystem) Download(ctx context.Context, link *Link) (*FileReader, error) {
	self.events.TriggerUpdate()

	link = self.links.LinkFromID(link.ID())
	if link == nil {
		return nil, ErrInvalidLink
	}

	if !link.IsFile() {
		return nil, ErrInvalidLinkType
	}

	share := link.Share()

	revision, err := self.client.GetRevisionAllBlocks(ctx, share.ID(), link.ID(), link.RevisionID())
	if err != nil {
		return nil, err
	}

	return &FileReader{
		ctx:    ctx,
		client: self.client,
		user:   self.user,
		link:   link,
		blocks: revision.Blocks,
	}, nil
}

func (self *FileSystem) Move(ctx context.Context, link *Link, parent *Link, name string) error {
	self.events.TriggerUpdate()

	// Make sure the links are up-to-date
	link = self.links.LinkFromID(link.ID())
	parent = self.links.LinkFromID(parent.ID())

	if link == nil || parent == nil {
		return ErrInvalidLink
	}

	share := link.Share()
	address := share.Address()
	srcParent := link.Parent()

	request := proton.MoveLinkReq{
		ParentLinkID:     parent.ID(),
		OriginalHash:     link.Hash(),
		SignatureAddress: address.Email(),
	}

	err := request.SetName(name, address.Keyring(), parent.Keyring())
	if err != nil {
		return err
	}

	err = request.SetHash(name, parent.HashKey())
	if err != nil {
		return err
	}

	nodePassphrase, err := reencryptKeyPacket(
		srcParent.Keyring(),
		parent.Keyring(),
		address.Keyring(),
		link.NodePassphrase(),
	)

	if err != nil {
		return err
	}

	request.NodePassphrase = nodePassphrase
	request.NodePassphraseSignature = link.NodePassphraseSignature()

	err = self.client.MoveLink(ctx, share.ID(), link.ID(), request)
	if err != nil {
		return err
	}

	self.events.TriggerUpdate()
	return nil
}

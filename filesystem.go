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

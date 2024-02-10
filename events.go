package drive

import (
	"context"
	"time"

	"github.com/henrybear327/go-proton-api"
)

type EventLoop struct {
	//
	// PARAMETERS
	//

	client *proton.Client
	links  *Links

	//
	// INTERNAL STATE
	//

	nextEvent string

	triggerUpdate chan struct{}
	waitUpdate    chan struct{}
}

func (self *EventLoop) Init(ctx context.Context) error {
	self.triggerUpdate = make(chan struct{})
	self.waitUpdate = make(chan struct{})

	err := self.getNextEvent(ctx)
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()

		for {
			externalTrigger := false

			select {
			case <-ctx.Done():
				return
			case <-self.triggerUpdate:
				externalTrigger = true
			case <-ticker.C:
				// ...
			}

			share := self.links.Share()

			for {
				event, err := self.client.GetShareEvent(ctx, share.ID(), self.nextEvent)
				if err != nil {
					continue
				}

				if len(event.Events) > 0 {
					_ = self.handleEvents(event.Events)
					self.nextEvent = event.EventID
				}

				if event.Refresh {
					_ = self.getNextEvent(ctx)
				}

				if len(event.Events) == 0 {
					break
				}
			}

			if externalTrigger {
				self.waitUpdate <- struct{}{}
			}
		}
	}()

	return nil
}

func (self *EventLoop) getNextEvent(ctx context.Context) error {
	share := self.links.Share()

	eventID, err := self.client.GetLatestShareEventID(ctx, share.ID())
	if err != nil {
		return err
	}

	self.nextEvent = eventID
	return nil
}

func (self *EventLoop) handleEvents(events []proton.LinkEvent) error {
	for _, event := range events {
		var err error = nil

		switch event.EventType {
		case proton.LinkEventCreate:
			err = self.links.OnCreate(event)
		case proton.LinkEventUpdate:
			err = self.links.OnUpdate(event)
		case proton.LinkEventUpdateMetadata:
			err = self.links.OnUpdate(event)
		default:
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (self *EventLoop) TriggerUpdate() {
	self.triggerUpdate <- struct{}{}
	<-self.waitUpdate
}

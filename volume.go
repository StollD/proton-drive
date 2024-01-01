package drive

import (
	"github.com/henrybear327/go-proton-api"
)

type Volume struct {
	volume proton.Volume
}

func (self *Volume) ID() string {
	return self.volume.VolumeID
}

func (self *Volume) ShareID() string {
	return self.volume.Share.ShareID
}

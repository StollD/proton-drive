package drive

import "time"

type Attributes struct {
	Size       int64
	Hash       string
	MIMEType   string
	BlockSizes []int64
	ModifyTime time.Time
}

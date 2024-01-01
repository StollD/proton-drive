package drive

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/henrybear327/go-proton-api"
	"io"
)

var (
	ErrOutOfRange              = errors.New("out of range read")
	ErrBlockAddressNotFound    = errors.New("block signature address not found")
	ErrBlockVerificationFailed = errors.New("block verification failed")
	ErrInvalidSeekOperation    = errors.New("invalid seek operation")
)

var _ io.Reader = &FileReader{}
var _ io.Seeker = &FileReader{}
var _ io.Closer = &FileReader{}

type FileReader struct {
	//
	// PARAMETERS
	//

	ctx context.Context

	client *proton.Client
	user   *User
	link   *Link

	blocks []proton.Block

	//
	// INTERNAL STATE
	//

	blockIndex  int
	blockOffset int64
	blockData   *bytes.Reader

	streamOffset int64
}

func (self *FileReader) Read(buffer []byte) (int, error) {
	err := self.updateCurrentBlock()
	if err != nil {
		return 0, err
	}

	_, err = self.blockData.Seek(self.streamOffset-self.blockOffset, io.SeekStart)
	if err != nil {
		return 0, err
	}

	n, err := self.blockData.Read(buffer)
	if err == nil {
		self.streamOffset += int64(n)
	}

	return n, err
}

func (self *FileReader) updateCurrentBlock() error {
	sizes := self.link.BlockSizes()

	index := -1
	currentOffset := int64(0)

	if self.streamOffset < 0 {
		return ErrOutOfRange
	}

	for i := 0; i < len(sizes); i++ {
		start := currentOffset
		end := currentOffset + sizes[i]

		if self.streamOffset >= start && self.streamOffset < end {
			index = i
			break
		}

		currentOffset = end
	}

	if index == -1 {
		return io.EOF
	}

	if self.blockIndex == index && self.blockData != nil {
		return nil
	}

	block := self.blocks[index]

	address := self.user.AddressFromEmail(block.SignatureEmail)
	if address == nil {
		return ErrBlockAddressNotFound
	}

	reader, err := self.client.GetBlock(self.ctx, block.BareURL, block.Token)
	if err != nil {
		return err
	}

	defer func() {
		_ = reader.Close()
	}()

	encrypted, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	hash := sha256.New()
	hash.Write(encrypted)

	if block.Hash != base64.StdEncoding.EncodeToString(hash.Sum(nil)) {
		return ErrBlockVerificationFailed
	}

	decrypted, err := self.link.SessionKey().Decrypt(encrypted)
	if err != nil {
		return err
	}

	signature, err := crypto.NewPGPMessageFromArmored(block.EncSignature)
	if err != nil {
		return err
	}

	err = address.Keyring().VerifyDetachedEncrypted(decrypted, signature, self.link.Keyring(), crypto.GetUnixTime())
	if err != nil {
		return err
	}

	self.blockIndex = index
	self.blockOffset = currentOffset
	self.blockData = bytes.NewReader(decrypted.GetBinary())

	return nil
}

func (self *FileReader) Size() int64 {
	var size int64 = 0

	for _, bs := range self.link.BlockSizes() {
		size += bs
	}

	return size
}

func (self *FileReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64 = 0

	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = self.streamOffset + offset
	case io.SeekEnd:
		abs = self.Size() + offset
	default:
		return 0, ErrInvalidSeekOperation
	}

	if abs < 0 {
		return 0, ErrInvalidSeekOperation
	}

	self.streamOffset = abs
	return abs, nil
}

func (self *FileReader) Close() error {
	self.blockData = nil
	self.blocks = nil
	self.streamOffset = 0

	return nil
}

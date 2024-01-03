package drive

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/henrybear327/go-proton-api"
	"hash"
	"io"
	"time"
)

const (
	BlockSize     = 4 * 1024 * 1024
	ISO8601Layout = "2006-01-02T15:04:05-0700"
)

var (
	ErrUnexpectedBlockUploadLinks = errors.New("unexpected number of block upload links")
)

var _ io.Writer = &FileWriter{}
var _ io.Closer = &FileWriter{}

type FileWriter struct {
	//
	// PARAMETERS
	//

	ctx context.Context

	client *proton.Client
	events *EventLoop

	parent     *Link
	linkID     string
	revisionID string

	newFile bool

	keyring    *crypto.KeyRing
	sessionKey *crypto.SessionKey

	//
	// INTERNAL STATE
	//

	blockIndex int
	blockSize  int
	blockData  []byte

	blockSizes  []int64
	blockHashes []byte

	contentSize int64
	contentHash hash.Hash
}

func (self *FileWriter) allocateState() {
	if self.blockData != nil {
		return
	}

	self.blockData = make([]byte, BlockSize)
	self.blockSizes = []int64{}
	self.blockHashes = []byte{}

	self.contentSize = 0
	self.contentHash = sha1.New()
}

func (self *FileWriter) Write(buffer []byte) (int, error) {
	self.allocateState()

	self.contentSize += int64(len(buffer))
	self.contentHash.Write(buffer)

	for i := 0; i < len(buffer); {
		missing := BlockSize - self.blockSize
		available := min(missing, len(buffer))

		start := self.blockSize
		end := start + available

		copy(self.blockData[start:end], buffer[i:available])

		i += available
		self.blockSize += available

		if available < missing {
			break
		}

		err := self.uploadCurrentBlock()
		if err != nil {
			return i, self.handleError(err)
		}

		self.blockIndex++
		self.blockSize = 0
	}

	return len(buffer), nil
}

func (self *FileWriter) uploadCurrentBlock() error {
	share := self.parent.Share()
	data := self.blockData[:self.blockSize]

	address := share.Address()
	message := crypto.NewPlainMessage(data)

	encrypted, err := self.sessionKey.Encrypt(message)
	if err != nil {
		return err
	}

	signature, err := address.Keyring().SignDetachedEncrypted(message, self.keyring)
	if err != nil {
		return err
	}

	signatureArm, err := signature.GetArmored()
	if err != nil {
		return err
	}

	blockHash := sha256.New()
	blockHash.Write(encrypted)

	checkSum := blockHash.Sum(nil)
	checkSumEnc := base64.StdEncoding.EncodeToString(checkSum)

	request := proton.BlockUploadReq{
		AddressID:  address.ID(),
		ShareID:    share.ID(),
		LinkID:     self.linkID,
		RevisionID: self.revisionID,

		BlockList: []proton.BlockUploadInfo{
			{
				Index:        self.blockIndex + 1,
				Size:         int64(len(encrypted)),
				Hash:         checkSumEnc,
				EncSignature: signatureArm,
			},
		},
	}

	rsp, err := self.client.RequestBlockUpload(self.ctx, request)
	if err != nil {
		return err
	}

	if len(rsp) != 1 {
		return ErrUnexpectedBlockUploadLinks
	}

	err = self.client.UploadBlock(self.ctx, rsp[0].BareURL, rsp[0].Token, bytes.NewReader(encrypted))
	if err != nil {
		return err
	}

	self.blockSizes = append(self.blockSizes, int64(self.blockSize))
	self.blockHashes = append(self.blockHashes, checkSum...)

	return nil
}

func (self *FileWriter) Close() error {
	self.allocateState()

	err := self.uploadCurrentBlock()
	if err != nil {
		return self.handleError(err)
	}

	share := self.parent.Share()
	address := share.Address()

	signature, err := address.Keyring().SignDetached(crypto.NewPlainMessage(self.blockHashes))
	if err != nil {
		return self.handleError(err)
	}

	signatureString, err := signature.GetArmored()
	if err != nil {
		return self.handleError(err)
	}

	request := proton.CommitRevisionReq{
		ManifestSignature: signatureString,
		SignatureAddress:  address.Email(),
	}

	xAttr := proton.RevisionXAttrCommon{
		Size:             self.contentSize,
		BlockSizes:       self.blockSizes,
		ModificationTime: time.Now().Format(ISO8601Layout),
		Digests: map[string]string{
			"SHA1": self.Hash(),
		},
	}

	err = request.SetEncXAttrString(address.Keyring(), self.keyring, &xAttr)
	if err != nil {
		return self.handleError(err)
	}

	err = self.client.CommitRevision(self.ctx, share.ID(), self.linkID, self.revisionID, request)
	if err != nil {
		return self.handleError(err)
	}

	self.blockData = nil
	self.blockSizes = nil
	self.blockHashes = nil

	self.events.TriggerUpdate()
	return nil
}

func (self *FileWriter) handleError(err error) error {
	share := self.parent.Share()

	if self.newFile {
		_ = self.client.DeleteChildren(self.ctx, share.ID(), self.parent.ID(), self.linkID)
	} else {
		_ = self.client.DeleteRevision(self.ctx, share.ID(), self.linkID, self.revisionID)
	}

	return err
}

func (self *FileWriter) Size() int64 {
	return self.contentSize
}

func (self *FileWriter) Hash() string {
	self.allocateState()
	return hex.EncodeToString(self.contentHash.Sum(nil))
}

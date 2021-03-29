package id

import (
	"fmt"
	"github.com/google/uuid"
)

const (
	Version1 = 1
	Version2 = 2
	Version3 = 3
	Version4 = 4
	Version5 = 5
)

// only version3&5 need data
// version1 can be used on distributed scenario
// version2 less to use
// version3 based on MD5 but less to use
// version4 based on random
// version5 based on SHA1 safer but slow
func NewUUID(version int, data []byte) (uuid.UUID, error) {
	switch version {
	case Version1:
		return uuid.NewUUID()
	case Version2:
		return uuid.NewDCEGroup()
	case Version3:
		return uuid.NewMD5(uuid.NameSpaceDNS, data), nil
	case Version4:
		return uuid.NewRandom()
	case Version5:
		return uuid.NewSHA1(uuid.NameSpaceOID, data), nil
	default:
		return uuid.UUID{}, fmt.Errorf("you must choose the version number in [1, 5]")
	}
}

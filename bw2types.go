package bw2bind

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/immesys/bw2/objects"
)

type PublishParams struct {
	URI                string
	PrimaryAccessChain string
	RoutingObjects     []objects.RoutingObject
	PayloadObjects     []PayloadObject
	Expiry             *time.Time
	ExpiryDelta        *time.Duration
	ElaboratePAC       string
	DoVerify           bool
	Persist            bool
}
type SubscribeParams struct {
	URI                string
	PrimaryAccessChain string
	RoutingObjects     []objects.RoutingObject
	Expiry             *time.Time
	ExpiryDelta        *time.Duration
	ElaboratePAC       string
	DoVerify           bool
	LeavePacked        bool
}
type ListParams struct {
	MVK                []byte
	URISuffix          string
	PrimaryAccessChain *objects.DChain
	RoutingObjects     []objects.RoutingObject
	Expiry             *time.Time
	ExpiryDelta        *time.Duration
	ElaboratePAC       int
	DoVerify           bool
}
type QueryParams struct {
	MVK                []byte
	URISuffix          string
	PrimaryAccessChain *objects.DChain
	RoutingObjects     []objects.RoutingObject
	Expiry             *time.Time
	ExpiryDelta        *time.Duration
	ElaboratePAC       int
	DoVerify           bool
}
type CreateDOTParams struct {
	IsPermission     bool
	To               string
	TTL              uint8
	Expiry           *time.Time
	ExpiryDelta      *time.Duration
	Contact          string
	Comment          string
	Revokers         []string
	OmitCreationDate bool

	//For Access
	URI               string
	AccessPermissions string

	//For Permissions
	AppPermissions map[string]string
}
type CreateDotChainParams struct {
	DOTs         []string
	IsPermission bool
	UnElaborate  bool
}
type CreateEntityParams struct {
	Expiry           *time.Time
	ExpiryDelta      *time.Duration
	Contact          string
	Comment          string
	Revokers         []string
	OmitCreationDate bool
}

type SimpleMessage struct {
	From     string
	URI      string
	POs      []PayloadObject
	ROs      []objects.RoutingObject
	POErrors []error
}

func (sm *SimpleMessage) Dump() {
	fmt.Printf("Message from %s on %s:", sm.From, sm.URI)
	for _, po := range sm.POs {
		fmt.Println(po.TextRepresentation())
	}
}
func PONumDotForm(ponum int) string {
	return fmt.Sprintf("%d.%d.%d.%d", ponum>>24, (ponum>>16)&0xFF, (ponum>>8)&0xFF, ponum&0xFF)
}
func PONumFromDotForm(dotform string) (int, error) {
	parts := strings.Split(dotform, ".")
	if len(parts) != 4 {
		return 0, errors.New("Bad dotform")
	}
	rv := 0
	for i := 0; i < 4; i++ {
		cx, err := strconv.ParseUint(parts[i], 10, 8)
		if err != nil {
			return 0, err
		}
		rv += (int(cx)) << uint(((3 - i) * 8))
	}
	return rv, nil
}

func FromDotForm(dotform string) int {
	rv, err := PONumFromDotForm(dotform)
	if err != nil {
		panic(err)
	}
	return rv
}

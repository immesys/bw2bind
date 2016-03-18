package bw2bind

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/immesys/bw2/objects"
)

const ElaborateDefault = ""
const ElaborateFull = "full"
const ElaboratePartial = "partial"
const ElaborateNone = "none"

type PublishParams struct {
	URI                string
	PrimaryAccessChain string
	AutoChain          bool
	RoutingObjects     []objects.RoutingObject
	PayloadObjects     []PayloadObject
	Expiry             *time.Time
	ExpiryDelta        *time.Duration
	ElaboratePAC       string
	DoNotVerify        bool
	Persist            bool
}
type SubscribeParams struct {
	URI                string
	PrimaryAccessChain string
	AutoChain          bool
	RoutingObjects     []objects.RoutingObject
	Expiry             *time.Time
	ExpiryDelta        *time.Duration
	ElaboratePAC       string
	DoNotVerify        bool
	LeavePacked        bool
}
type ListParams struct {
	URI                string
	PrimaryAccessChain string
	AutoChain          bool
	RoutingObjects     []objects.RoutingObject
	Expiry             *time.Time
	ExpiryDelta        *time.Duration
	ElaboratePAC       string
	DoNotVerify        bool
}
type QueryParams struct {
	URI                string
	PrimaryAccessChain string
	AutoChain          bool
	RoutingObjects     []objects.RoutingObject
	Expiry             *time.Time
	ExpiryDelta        *time.Duration
	ElaboratePAC       string
	DoNotVerify        bool
	LeavePacked        bool
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
type BuildChainParams struct {
	URI         string
	Permissions string
	To          string
}
type SimpleMessage struct {
	From     string
	URI      string
	POs      []PayloadObject
	ROs      []objects.RoutingObject
	POErrors []error
}
type SimpleChain struct {
	Hash        string
	Permissions string
	URI         string
	To          string
}

func (sm *SimpleMessage) Dump() {
	fmt.Printf("Message from %s on %s:\n", sm.From, sm.URI)
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

func (sm *SimpleMessage) GetOnePODF(df string) PayloadObject {
	for _, p := range sm.POs {
		if p.IsTypeDF(df) {
			return p
		}
	}
	return nil
}

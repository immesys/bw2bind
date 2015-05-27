package bw2bind

import (
	"time"

	"github.com/immesys/bw2/objects"
)

type PublishParams struct {
	URI                string
	PrimaryAccessChain string
	RoutingObjects     []objects.RoutingObject
	PayloadObjects     []objects.PayloadObject
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
	From string
	URI  string
	POs  []objects.PayloadObject
	ROs  []objects.RoutingObject
}

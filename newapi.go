package bw2bind

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/immesys/bw2/crypto"
	"github.com/immesys/bw2/objects"
)

func (cl *BW2Client) PublishDOTWithAcc(blob []byte, account int) (string, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdPutDot, seqno)
	//Strip first byte of blob, assuming it came from a file
	po := CreateBasePayloadObject(PONumROAccessDOT, blob)
	req.AddPayloadObject(po)
	req.AddHeader("account", strconv.Itoa(account))
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", err
	}
	hash, _ := fr.GetFirstHeader("hash")
	return hash, nil

}
func (cl *BW2Client) PublishDOT(blob []byte) (string, error) {
	return cl.PublishDOTWithAcc(blob, 0)
}
func (cl *BW2Client) PublishEntityWithAcc(blob []byte, account int) (string, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdPutEntity, seqno)
	po := CreateBasePayloadObject(PONumROEntity, blob)
	req.AddPayloadObject(po)
	req.AddHeader("account", strconv.Itoa(account))
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", err
	}
	vk, _ := fr.GetFirstHeader("vk")
	return vk, nil

}
func (cl *BW2Client) PublishEntity(blob []byte) (string, error) {
	return cl.PublishEntityWithAcc(blob, 0)
}
func (cl *BW2Client) PublishChainWithAcc(blob []byte, account int) (string, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdPutChain, seqno)
	//TODO it might not be with a key...
	po := CreateBasePayloadObject(PONumROAccessDChain, blob)
	req.AddPayloadObject(po)
	req.AddHeader("account", strconv.Itoa(account))
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", err
	}
	hash, _ := fr.GetFirstHeader("hash")
	return hash, nil

}
func (cl *BW2Client) PublishChain(blob []byte) (string, error) {
	return cl.PublishChainWithAcc(blob, 0)
}
func (cl *BW2Client) ResolveLongAlias(al string) (data []byte, zero bool, err error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdResolveAlias, seqno)
	req.AddHeader("longkey", al)
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return nil, false, err
	}
	v, _ := fr.GetFirstHeaderB("value")
	return v, bytes.Equal(v, make([]byte, 32)), nil
}
func (cl *BW2Client) ResolveShortAlias(al string) (data []byte, zero bool, err error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdResolveAlias, seqno)
	req.AddHeader("shortkey", al)
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return nil, false, err
	}
	v, _ := fr.GetFirstHeaderB("value")
	return v, bytes.Equal(v, make([]byte, 32)), nil
}
func (cl *BW2Client) ResolveEmbeddedAlias(al string) (data string, err error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdResolveAlias, seqno)
	req.AddHeader("longkey", al)
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", err
	}
	v, _ := fr.GetFirstHeader("value")
	return v, nil
}

type RegistryValidity int

const (
	StateUnknown = iota
	StateValid
	StateExpired
	StateRevoked
	StateError
)

func (cl *BW2Client) ValidityToString(i RegistryValidity, err error) string {
	switch i {
	case StateUnknown:
		return "UNKNOWN"
	case StateValid:
		return "valid"
	case StateExpired:
		return "EXPIRED"
	case StateRevoked:
		return "REVOKED"
	case StateError:
		return "ERROR: " + err.Error()
	}
	return "<WTF?>"
}
func (cl *BW2Client) ResolveRegistry(key string) (ro objects.RoutingObject, validity RegistryValidity, err error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdResolveRegistryObject, seqno)
	req.AddHeader("key", key)
	fr := <-cl.transact(req)
	if er := fr.MustResponse(); er != nil {
		return nil, StateError, er
	}
	if len(fr.GetAllROs()) == 0 {
		return nil, StateUnknown, nil
	}
	ro = fr.GetAllROs()[0]
	err = nil
	valid, _ := fr.GetFirstHeader("validity")
	switch valid {
	case "valid":
		validity = StateValid
		return
	case "expired":
		validity = StateExpired
		return
	case "revoked":
		validity = StateRevoked
		return
	default:
		panic(valid)
	}
}

type BalanceInfo struct {
	Addr    string
	Human   string
	Decimal string
	Int     *big.Int
}

func (cl *BW2Client) EntityBalances() ([]*BalanceInfo, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdEntityBalances, seqno)
	fr := <-cl.transact(req)
	if er := fr.MustResponse(); er != nil {
		return nil, er
	}
	rv := make([]*BalanceInfo, 0, 16)
	for _, poe := range fr.POs {
		if poe.PONum == PONumAccountBalance {
			parts := strings.Split(string(poe.PO), ",")
			i := big.NewInt(0)
			i, _ = i.SetString(parts[1], 10)
			rv = append(rv, &BalanceInfo{
				Addr:    parts[0],
				Decimal: parts[1],
				Human:   parts[2],
				Int:     i,
			})
		}
	}
	return rv, nil
}
func (cl *BW2Client) AddressBalance(addr string) (*BalanceInfo, error) {
	if addr[0:2] == "0x" {
		addr = addr[2:]
	}
	if len(addr) != 40 {
		return nil, fmt.Errorf("Address must be 40 hex characters")
	}
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdAddressBalance, seqno)
	req.AddHeader("address", addr)
	fr := <-cl.transact(req)
	if er := fr.MustResponse(); er != nil {
		return nil, er
	}
	poe := fr.POs[0]
	parts := strings.Split(string(poe.PO), ",")
	i := big.NewInt(0)
	i, _ = i.SetString(parts[1], 10)
	rv := &BalanceInfo{
		Addr:    parts[0],
		Decimal: parts[1],
		Human:   parts[2],
		Int:     i,
	}
	return rv, nil
}

type BCIP struct {
	Confirmations *int64
	Timeout       *int64
	Maxage        *int64
}

type CurrentBCIP struct {
	Confirmations int64
	Timeout       int64
	Maxage        int64
	CurrentAge    time.Duration
	CurrentBlock  uint64
	Peers         int64
	HighestBlock  int64
	Difficulty    int64
}

func (cl *BW2Client) GetBCInteractionParams() (*CurrentBCIP, error) {
	return cl.SetBCInteractionParams(nil)
}
func (cl *BW2Client) SetBCInteractionParams(to *BCIP) (*CurrentBCIP, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdBCInteractionParams, seqno)
	if to != nil {
		if to.Confirmations != nil {
			req.AddHeader("confirmations", strconv.FormatInt(*to.Confirmations, 10))
		}
		if to.Timeout != nil {
			req.AddHeader("timeout", strconv.FormatInt(*to.Timeout, 10))
		}
		if to.Maxage != nil {
			req.AddHeader("maxage", strconv.FormatInt(*to.Maxage, 10))
		}
	}
	fr := <-cl.transact(req)
	if er := fr.MustResponse(); er != nil {
		return nil, er
	}
	rv := &CurrentBCIP{}
	v, _ := fr.GetFirstHeader("confirmations")
	iv, _ := strconv.ParseInt(v, 10, 64)
	rv.Confirmations = iv
	v, _ = fr.GetFirstHeader("timeout")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.Timeout = iv
	v, _ = fr.GetFirstHeader("maxage")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.Maxage = iv
	v, _ = fr.GetFirstHeader("currentblock")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.CurrentBlock = uint64(iv)
	v, _ = fr.GetFirstHeader("currentage")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.CurrentAge = time.Duration(iv) * time.Second
	v, _ = fr.GetFirstHeader("peers")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.Peers = iv
	v, _ = fr.GetFirstHeader("highest")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.HighestBlock = iv
	v, _ = fr.GetFirstHeader("difficulty")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.Difficulty = iv
	return rv, nil
}

type Currency int64

const KiloEther Currency = 1000 * Ether
const Ether Currency = 1000 * MilliEther
const MilliEther Currency = 1000 * MicroEther
const Finney Currency = 1000 * MicroEther
const MicroEther Currency = 1000 * NanoEther
const Szabo Currency = 1000 * NanoEther
const NanoEther Currency = 1
const GigaWei Currency = 1

func CurrencyToWei(v Currency) *big.Int {
	rv := big.NewInt(int64(v))
	rv = rv.Mul(rv, big.NewInt(1000000000))
	return rv
}

func (cl *BW2Client) TransferWei(from int, to string, wei *big.Int) error {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdTransfer, seqno)
	req.AddHeader("account", strconv.Itoa(from))
	req.AddHeader("address", to)
	req.AddHeader("valuewei", wei.Text(10))
	return (<-cl.transact(req)).MustResponse()
}
func (cl *BW2Client) TransferFrom(from int, to string, value Currency) error {
	return cl.TransferWei(from, to, CurrencyToWei(value))
}
func (cl *BW2Client) Transfer(to string, value Currency) error {
	return cl.TransferFrom(0, to, value)
}
func (cl *BW2Client) NewDesignatedRouterOffer(account int, nsvk string, dr *objects.Entity) error {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdNewDROffer, seqno)
	req.AddHeader("account", strconv.Itoa(account))
	req.AddHeader("nsvk", nsvk)
	if dr != nil {
		po := CreateBasePayloadObject(objects.ROEntityWKey, dr.GetSigningBlob())
		req.AddPayloadObject(po)
	}
	return (<-cl.transact(req)).MustResponse()
}

func (cl *BW2Client) GetDesignatedRouterOffers(nsvk string) (active string, activesrv string, drvks []string, err error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdListDROffers, seqno)
	req.AddHeader("nsvk", nsvk)
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", "", nil, err
	}
	rv := make([]string, 0)
	for _, po := range fr.POs {
		if po.PONum == objects.RODesignatedRouterVK {
			rv = append(rv, crypto.FmtKey(po.PO))
		}
	}
	act, _ := fr.GetFirstHeader("active")
	srv, _ := fr.GetFirstHeader("srv")
	return act, srv, rv, nil
}
func (cl *BW2Client) AcceptDesignatedRouterOffer(account int, drvk string, ns *objects.Entity) error {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdAcceptDROffer, seqno)
	req.AddHeader("account", strconv.Itoa(account))
	req.AddHeader("drvk", drvk)
	if ns != nil {
		po := CreateBasePayloadObject(objects.ROEntityWKey, ns.GetSigningBlob())
		req.AddPayloadObject(po)
	}
	return (<-cl.transact(req)).MustResponse()
}
func (cl *BW2Client) SetDesignatedRouterSRVRecord(account int, srv string, dr *objects.Entity) error {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdUpdateSRVRecord, seqno)
	req.AddHeader("account", strconv.Itoa(account))
	req.AddHeader("srv", srv)
	if dr != nil {
		po := CreateBasePayloadObject(objects.ROEntityWKey, dr.GetSigningBlob())
		req.AddPayloadObject(po)
	}
	return (<-cl.transact(req)).MustResponse()
}
func (cl *BW2Client) CreateLongAlias(account int, key []byte, val []byte) error {
	if len(key) > 32 || len(val) > 32 {
		return fmt.Errorf("Key and value must be shorter than 32 bytes")
	}
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdMakeLongAlias, seqno)
	req.AddHeader("account", strconv.Itoa(account))
	req.AddHeaderB("content", val)
	req.AddHeaderB("key", key)
	return (<-cl.transact(req)).MustResponse()
}

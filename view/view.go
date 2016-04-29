package view

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/immesys/bw2/util"
	"github.com/immesys/bw2bind"
)

type View struct {
	cl        *bw2bind.BW2Client
	ex        Expression
	metastore map[string]map[string]*bw2bind.MetadataTuple
	ns        []string
	msmu      sync.RWMutex
	mscond    *sync.Cond
	msloaded  bool
	changecb  []func()
}

func (v *View) Meta(uri, key string) (*bw2bind.MetadataTuple, bool) {
	//This will check uri, and parents (meta is inherited)
	parts := strings.Split(uri, "/")
	var val *bw2bind.MetadataTuple = nil
	set := false
	v.msmu.RLock()
	for i := 1; i <= len(parts); i++ {
		uri := strings.Join(parts[:i], "/")
		m1, ok := v.metastore[uri]
		if ok {
			v, subok := m1[key]
			if subok {
				val = v
				set = true
			}
		}
	}
	v.msmu.RUnlock()
	return val, set
}

func (v *View) AllMeta(uri string) map[string]*bw2bind.MetadataTuple {
	parts := strings.Split(uri, "/")
	rv := make(map[string]*bw2bind.MetadataTuple)
	v.msmu.RLock()
	for i := 1; i <= len(parts); i++ {
		uri := strings.Join(parts[:i], "/")
		m1, ok := v.metastore[uri]
		if ok {
			for kk, vv := range m1 {
				rv[kk] = vv
			}
		}
	}
	v.msmu.RUnlock()
	return rv
}

type Expression interface {
	//Given a complete resource name, does this expression
	//permit it to be included in the view
	Matches(uri string, v *View) bool
	//Given a partial resource name (prefix) does this expression
	//possibly permit it to be included in the view. Yes means maybe
	//no means no.
	MightMatch(uri string, v *View) bool

	//Return a list of all URIs(sans namespaces) that are sufficient
	//to evaluate this expression (minimum subscription set). Does not
	//include metadata
	CanonicalSuffixes() []string
}
type andExpression struct {
	subex []Expression
}

func (e *andExpression) Matches(uri string, v *View) bool {
	for _, s := range e.subex {
		if !s.Matches(uri, v) {
			return false
		}
	}
	return true
}
func (e *andExpression) MightMatch(uri string, v *View) bool {
	for _, s := range e.subex {
		if !s.MightMatch(uri, v) {
			return false
		}
	}
	return true
}

/*
  (a or b) and (c or d)
*/

func foldAndCanonicalSuffixes(lhs []string, rhsz ...[]string) []string {
	if len(rhsz) == 0 {
		return lhs
	}

	rhs := rhsz[0]
	retv := []string{}
	for _, lv := range lhs {
		for _, rv := range rhs {
			res, ok := util.RestrictBy(lv, rv)
			if ok {
				retv = append(retv, res)
			}
		}
	}
	//Now we need to dedup RV
	// if A restrictBy B == A, then A is redundant and B is superior
	//                   == B, then B is redundant and A is superior
	//  is not equal to either, then both are relevant
	dedup := []string{}
	for out := 0; out < len(retv); out++ {
		for in := 0; in < len(retv); in++ {
			if in == out {
				continue
			}
			res, ok := util.RestrictBy(retv[out], retv[in])
			if ok {
				if res == retv[out] && retv[out] != retv[in] {
					//out is redundant to in, and they are not identical
					//do not add out, as we will add in later
					goto nextOut
				}
				if retv[out] == retv[in] {
					//they are identical (and reduandant) so only add
					//out if it is less than in
					if out > in {
						goto nextOut
					}
				}
			}
		}
		dedup = append(dedup, retv[out])
	nextOut:
	}
	return foldAndCanonicalSuffixes(dedup, rhsz[1:]...)
}
func (e *andExpression) CanonicalSuffixes() []string {
	retv := [][]string{}
	for _, s := range e.subex {
		retv = append(retv, s.CanonicalSuffixes())
	}
	return foldAndCanonicalSuffixes(retv[0], retv[1:]...)
}

type orExpression struct {
	subex []Expression
}

func (e *orExpression) Matches(uri string, v *View) bool {
	for _, s := range e.subex {
		if s.Matches(uri, v) {
			return true
		}
	}
	return false
}
func (e *orExpression) MightMatch(uri string, v *View) bool {
	for _, s := range e.subex {
		if s.MightMatch(uri, v) {
			return true
		}
	}
	return false
}
func (e *orExpression) CanonicalSuffixes() []string {
	retv := []string{}
	for _, s := range e.subex {
		retv = append(retv, s.CanonicalSuffixes()...)
	}
	return retv
}

type metaEqExpression struct {
	key   string
	val   string
	regex bool
}

func (e *metaEqExpression) Matches(uri string, v *View) bool {
	val, ok := v.Meta(uri, e.key)
	if !ok {
		return false
	}
	if e.regex {
		panic("have not done regex yet")
	} else {
		fmt.Println("returning meta match: ", val.Value == e.val)
		return val.Value == e.val
	}
}
func (e *metaEqExpression) MightMatch(uri string, v *View) bool {
	//You don't know until the final resource
	return true
}
func (e *metaEqExpression) CanonicalSuffixes() []string {
	return []string{"*"}
}

type uriEqExpression struct {
	pattern string
	regex   bool
}

func (e *uriEqExpression) Matches(uri string, v *View) bool {
	if e.regex {
		return regexp.MustCompile(e.pattern).MatchString(uri)
	} else {
		panic("have not done thing yet")
	}
}
func (e *uriEqExpression) MightMatch(uri string, v *View) bool {
	if e.regex {
		//I'm sure we can change this in future, but it is hard
		return true
	} else {
		rhs := strings.Split(uri, "/")
		lhs := strings.Split(e.pattern, "/")
		//First check if NS matches (if present)
		if lhs[0] != "" {
			if rhs[0] != lhs[0] {
				return false
			}
		}
		li := 1
		ri := 1
		for li < len(lhs) && ri < len(rhs) {
			if lhs[li] == "*" {
				//Can arbitrarily expand
				return true
			}
			if lhs[li] == "+" ||
				lhs[li] == rhs[li] {
				li++
				ri++
				continue
			}
			return false
		}
		//either lhs or rhs is finished
		if li == len(lhs) {
			//Won't match, no more room in lhs pattern
			return false
		}
		//but if rhs finished we don't know
		return true
	}
}
func (e *uriEqExpression) CanonicalSuffixes() []string {
	if e.regex {
		return []string{"*"}
	}
	return []string{e.pattern}
}

func And(terms ...Expression) Expression {
	return &andExpression{subex: terms}
}
func Or(terms ...Expression) Expression {
	return &orExpression{subex: terms}
}
func EqMeta(key, value string) Expression {
	return &metaEqExpression{key: key, val: value, regex: false}
}
func RegexURI(pattern string) Expression {
	return &uriEqExpression{pattern: pattern, regex: true}
}

//If the URI does not begin with a slash it is considered a full
//uri. If it begins with a slash it has an implicit namespace filled
//in with the namespaces from NewView
func MatchURI(pattern string) Expression {
	return &uriEqExpression{pattern: pattern, regex: false}
}
func Prefix(pattern string) Expression {
	if pattern[len(pattern)-1] != '/' {
		pattern = pattern + "/"
	}
	return MatchURI(pattern + "*")
}
func Service(name string) Expression {
	//uri is .../service/selector/interface/sigslot/endpoint
	return MatchURI(fmt.Sprintf("/*/%s/+/+/+/+", name))
}
func Interface(name string) Expression {
	return RegexURI("^.*/" + name + "$")
}
func NewView(cl *bw2bind.BW2Client, namespaces []string, exz ...Expression) *View {
	ex := And(exz...)
	nsmap := make(map[string]struct{})
	for _, i := range namespaces {
		parts := strings.Split(i, "/")
		nsmap[parts[0]] = struct{}{}
	}
	ns := make([]string, 0, len(nsmap))
	for k, _ := range nsmap {
		ns = append(ns, k)
	}
	rv := &View{
		cl:        cl,
		ex:        ex,
		metastore: make(map[string]map[string]*bw2bind.MetadataTuple),
		ns:        ns,
	}
	rv.initMetaView()
	rv.waitForMetaView()
	return rv
}

func (v *View) waitForMetaView() {
	v.msmu.Lock()
	for !v.msloaded {
		v.mscond.Wait()
	}
	v.msmu.Unlock()
}
func (v *View) fireChange() {
	v.msmu.RLock()
	for _, cb := range v.changecb {
		go cb()
	}
	v.msmu.RUnlock()
}
func (v *View) initMetaView() {
	v.mscond = sync.NewCond(&v.msmu)
	procChange := func(sm *bw2bind.SimpleMessage) {
		fmt.Println("doing procchange")
		groups := regexp.MustCompile("^(.*)/!meta/([^/]*)$").FindStringSubmatch(sm.URI)
		if groups == nil {
			panic("bad re match")
		}
		uri := groups[1]
		key := groups[2]
		v.msmu.Lock()
		map1, ok := v.metastore[uri]
		if !ok {
			map1 = make(map[string]*bw2bind.MetadataTuple)
			v.metastore[uri] = map1
		}
		poi := sm.GetOnePODF(bw2bind.PODFSMetadata)
		if poi != nil {
			po := poi.(bw2bind.MetadataPayloadObject)
			map1[key] = po.Value()
		} else {
			delete(map1, key)
		}
		v.msmu.Unlock()
		v.fireChange()
	}
	go func() {
		//First subscribe and wait for that to finish
		rcz := make([]chan *bw2bind.SimpleMessage, len(v.ns))
		for i, n := range v.ns {
			fmt.Println("sub is on", n+"/*/!meta/+")
			rcz[i] = v.cl.SubscribeOrExit(&bw2bind.SubscribeParams{
				URI:       n + "/*/!meta/+",
				AutoChain: true,
			})
		}
		//Then we query
		for _, n := range v.ns {
			qres := v.cl.QueryOrExit(&bw2bind.QueryParams{
				URI:       n + "/*/!meta/+",
				AutoChain: true,
			})
			for m := range qres {
				if m == nil {
					panic("this is dumb")
				}
				procChange(m)
			}
		}
		//Then we mark store as populated
		v.msmu.Lock()
		v.msloaded = true
		v.msmu.Unlock()
		v.mscond.Broadcast()
		//And process our subscriptions
		for _, rch := range rcz {
			go func(r chan *bw2bind.SimpleMessage) {
				fmt.Println("starting listening on sub")
				for m := range r {
					fmt.Println("got a message")
					procChange(m)
				}
				fmt.Println("sub aborted")
			}(rch)
		}
	}()
}
func (v *View) PubSlot(iface, slot string, poz []bw2bind.PayloadObject) error {
	idz := v.Interfaces()
	fmt.Println("we found ", len(idz), "interfaces")
	fmt.Println(idz)
	for _, viewiface := range idz {
		if viewiface.Interface == iface {
			fmt.Println("doing the actual publish")
			err := v.cl.Publish(&bw2bind.PublishParams{
				AutoChain:      true,
				PayloadObjects: poz,
				URI:            viewiface.URI + "/slot/" + slot,
			})
			if err != nil {
				return err
			}
		} else {
			fmt.Println("iface mismatch", viewiface.Interface, iface)
		}
	}
	return nil
}
func (v *View) SubSignal(iface, signal string) (chan *bw2bind.SimpleMessage, error) {
	idz := v.Interfaces()
	rv := make(chan *bw2bind.SimpleMessage, 10)
	wg := sync.WaitGroup{}
	for _, viewiface := range idz {
		if viewiface.Interface == iface {
			rvc, err := v.cl.Subscribe(&bw2bind.SubscribeParams{
				AutoChain: true,
				URI:       viewiface.URI + "/signal/" + signal,
			})
			if err != nil {
				return nil, err
			}
			wg.Add(1)
			go func() {
				for sm := range rvc {
					rv <- sm
				}
				wg.Done()
			}()
		}
	}
	go func() {
		wg.Wait()
		close(rv)
	}()
	return rv, nil
}
func (v *View) Interfaces() []InterfaceDescription {
	v.msmu.RLock()
	found := make(map[string]InterfaceDescription)
	for uri, _ := range v.metastore {
		fmt.Println("checking ", uri)
		if v.ex.Matches(uri, v) {
			fmt.Println("passed first check")
			pat := `^(([^/]+)(/.*)?/(s\.[^/]+)/([^/]+)/(i\.[^/]+)).*$`
			//"^((([^/]+)/(.*)/(s\\.[^/]+)/+)/(i\\.[^/]+)).*$"
			groups := regexp.MustCompile(pat).FindStringSubmatch(uri)
			if groups != nil {
				id := InterfaceDescription{
					URI:       groups[1],
					Interface: groups[6],
					Service:   groups[4],
					Namespace: groups[2],
					Prefix:    groups[5],
					v:         v,
				}
				id.Suffix = strings.TrimPrefix(id.URI, id.Namespace+"/")
				fmt.Println("id was", id)
				found[id.URI] = id
			}
		}
	}
	v.msmu.RUnlock()
	rv := []InterfaceDescription{}
	for _, vv := range found {
		if vv.Meta("lastalive") != "" {
			rv = append(rv, vv)
		} else {
			fmt.Println("interface is not alive")
		}
	}
	sort.Sort(interfaceSorter(rv))
	return rv
}

type interfaceSorter []InterfaceDescription

func (is interfaceSorter) Swap(i, j int) {
	is[i], is[j] = is[j], is[i]
}
func (is interfaceSorter) Less(i, j int) bool {
	return strings.Compare(is[i].URI, is[j].URI) < 0
}
func (is interfaceSorter) Len() int {
	return len(is)
}
func (v *View) OnChange(f func()) {
	//TODO determien if change is actually relevant
	v.msmu.Lock()
	v.changecb = append(v.changecb, f)
	v.msmu.Unlock()
}

type InterfaceDescription struct {
	URI       string
	Interface string
	Service   string
	Namespace string
	Prefix    string
	Suffix    string
	v         *View
}

func (id *InterfaceDescription) Meta(key string) string {
	mdat, ok := id.v.Meta(id.URI, key)
	if !ok {
		return "<unset>"
	}
	return mdat.Value
}

/*
Example use
v := cl.NewView()
q := view.MatchURI(mypattern)
q = q.And(view.MetaEq(key, value))
q = q.And(view.MetaHasKey(key))
q = q.And(view.IsInterface("i.wavelet"))
q = q.And(view.IsService("s.thingy"))
v = v.And(view.MatchURI(myurl, mypattern))

can assume all interfaces are persisted up to .../i.foo/
when you subscribe,
*/

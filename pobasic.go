package bw2bind

import "fmt"

type POConstructor struct {
	PONum       string
	Mask        int
	Constructor func(int, []byte) (PayloadObject, error)
}

var PayloadObjectConstructors = []POConstructor{
	{"64.0.0.0", 4, LoadTextPayloadObjectPO},
	{"0.0.0.0", 0, LoadBasePayloadObjectPO},
}

func LoadPayloadObject(ponum int, contents []byte) (PayloadObject, error) {
	for _, c := range PayloadObjectConstructors {
		cponum, _ := PONumFromDotForm(c.PONum)
		cponum = cponum >> uint(32-c.Mask)
		if (ponum >> uint(32-c.Mask)) == cponum {
			return c.Constructor(ponum, contents)
		}
	}
	panic("Could not load PO")
}

//PayloadObject implements 0.0.0.0/0 : base
type PayloadObject interface {
	GetPONum() int
	GetPODotNum() string
	TextRepresentation() string
	GetContents() []byte
}
type PayloadObjectImpl struct {
	ponum    int
	contents []byte
}

func LoadBasePayloadObject(ponum int, contents []byte) (*PayloadObjectImpl, error) {
	return &PayloadObjectImpl{ponum: ponum, contents: contents}, nil
}
func LoadBasePayloadObjectPO(ponum int, contents []byte) (PayloadObject, error) {
	return LoadBasePayloadObject(ponum, contents)
}
func CreateBasePayloadObject(ponum int, contents []byte) *PayloadObjectImpl {
	rv, _ := LoadBasePayloadObject(ponum, contents)
	return rv
}
func (po *PayloadObjectImpl) GetPONum() int {
	return po.ponum
}
func (po *PayloadObjectImpl) SetPONum(ponum int) {
	po.ponum = ponum
}
func (po *PayloadObjectImpl) GetContents() []byte {
	return po.contents
}
func (po *PayloadObjectImpl) SetContents(v []byte) {
	po.contents = v
}
func (po *PayloadObjectImpl) GetPODotNum() string {
	return fmt.Sprintf("%d.%d.%d.%d", po.ponum>>24, (po.ponum>>16)&0xFF, (po.ponum>>8)&0xFF, po.ponum&0xFF)
}
func (po *PayloadObjectImpl) TextRepresentation() string {
	return fmt.Sprintf("PO %s len %d", PONumDotForm(po.ponum), len(po.contents))
}

//TextPayloadObject implements 64.0.0.0/4 : Human readable
type TextPayloadObject interface {
	PayloadObject
	Value() string
}
type TextPayloadObjectImpl struct {
	PayloadObjectImpl
}

func LoadTextPayloadObject(ponum int, contents []byte) (*TextPayloadObjectImpl, error) {
	bpl, _ := LoadBasePayloadObject(ponum, contents)
	return &TextPayloadObjectImpl{*bpl}, nil
}
func LoadTextPayloadObjectPO(ponum int, contents []byte) (PayloadObject, error) {
	return LoadTextPayloadObject(ponum, contents)
}
func CreateTextPayloadObject(ponum int, contents string) *TextPayloadObjectImpl {
	rv, _ := LoadTextPayloadObject(ponum, []byte(contents))
	return rv
}
func (po *TextPayloadObjectImpl) TextRepresentation() string {
	return fmt.Sprintf("PO %s len %d:\n%s", PONumDotForm(po.ponum), len(po.contents), string(po.contents))
}
func (po *TextPayloadObjectImpl) Value() string {
	return string(po.contents)
}

//StringPayloadObject implements 64.0.1.0/32 : String
func CreateStringPayloadObject(v string) TextPayloadObject {
	return CreateTextPayloadObject(FromDotForm("64.0.1.0"), v)
}

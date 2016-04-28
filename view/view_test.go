package view

import (
	"fmt"
	"reflect"
	"testing"
)

type testExpression1 struct {
	suffixes []string
}

func (t *testExpression1) CanonicalSuffixes() []string {
	return t.suffixes
}
func (t *testExpression1) Matches(uri string, v *View) bool {
	panic("")
}
func (t *testExpression1) MightMatch(uri string, v *View) bool {
	panic("")
}
func u(args ...string) []string {
	return args
}

type tc struct {
	lhs []string
	rhs []string
	res []string
}

func TestSuffixes(t *testing.T) {
	cases := []tc{
		tc{u("foo"), u("foo"), u("foo")},
		tc{u("foo/bar", "foo/baz"), u("foo"), u()},
		tc{u("foo/bar", "foo/+"), u("foo"), u()},
		tc{u("foo/bar", "foo/+"), u("foo/bop"), u("foo/bop")},
		tc{u("foo/bar", "foo/+"), u("*/bar"), u("foo/bar")},
	}
	for _, c := range cases {
		lhs := testExpression1{c.lhs}
		rhs := testExpression1{c.rhs}
		ae := And(&lhs, &rhs)
		z := ae.CanonicalSuffixes()
		if !(len(z) == 0 && len(c.res) == 0) && !reflect.DeepEqual(z, c.res) {
			fmt.Printf("for tc=%+v got z=%v\n", c, z)

			t.Fail()
		}
	}
}
func TestSuffixes2(t *testing.T) {
	x := And(Or(Iface("i.foobar"),
		Iface("i.foobaz")), Prefix("/foo/baz/"))
	fmt.Println(x.CanonicalSuffixes())
}

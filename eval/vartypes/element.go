package vartypes

import (
	"github.com/elves/elvish/eval/types"
)

type elem struct {
	variable Variable
	assocers []interface{}
	indices  []interface{}
	setValue interface{}
}

func (ev *elem) Set(v0 interface{}) error {
	var err error
	v := v0
	// Evaluate the actual new value from inside out. See comments in
	// MakeElement for how element assignment works.
	for i := len(ev.assocers) - 1; i >= 0; i-- {
		v, err = types.Assoc(ev.assocers[i], ev.indices[i], v)
		if err != nil {
			return err
		}
	}
	err = ev.variable.Set(v)
	// XXX(xiaq): Remember the set value for use in Get.
	ev.setValue = v0
	return err
}

func (ev *elem) Get() interface{} {
	// XXX(xiaq): This is only called from fixNilVariables. We don't want to
	// waste time accessing the variable, so we simply return the value that was
	// set.
	return ev.setValue
}

// NewElement returns an ephemeral variable used for assigning variable element.
func NewElement(v Variable, a []interface{}, i []interface{}) Variable {
	return &elem{v, a, i, ""}
}

// MakeElement returns a variable, that when set, simulates the mutation of an
// element.
func MakeElement(v Variable, indicies []interface{}) (Variable, error) {
	// Assignment of indexed variables actually assignes the variable, with
	// the right hand being a nested series of Assocs. As the simplest
	// example, `a[0] = x` is equivalent to `a = (assoc $a 0 x)`. A more
	// complex example is that `a[0][1][2] = x` is equivalent to
	//	`a = (assoc $a 0 (assoc $a[0] 1 (assoc $a[0][1] 2 x)))`.
	// Note that in each assoc form, the first two arguments can be
	// determined now, while the last argument is only known when the
	// right-hand-side is known. So here we evaluate the first two arguments
	// of each assoc form and put them in two slices, assocers and indicies.
	// In the previous example, the two slices will contain:
	//
	// assocers: $a $a[0] $a[0][1]
	// indicies:  0     1        2
	//
	// When the right-hand side of the assignment becomes available, the new
	// value for $a is evaluated by doing Assoc from inside out.
	assocers := make([]interface{}, len(indicies))
	varValue := v.Get()
	assocers[0] = varValue
	for i, index := range indicies[:len(indicies)-1] {
		lastAssocer := assocers[i]
		v, err := types.Index(lastAssocer, index)
		if err != nil {
			return nil, err
		}
		assocers[i+1] = v
	}
	return NewElement(v, assocers, indicies), nil
}

type elemErr struct {
	level int
	msg   string
}

func (err elemErr) Error() string {
	return err.msg
}

// GetHeadOfElement gets the underlying head variable of an element variable, or
// nil if the argument is not an element variable.
func GetHeadOfElement(v Variable) Variable {
	if ev, ok := v.(*elem); ok {
		return ev.variable
	}
	return nil
}

// GetElementErrorLevel returns the level of an error returned by MakeElement or
// DelElement. Level 0 represents that the error is about the variable itself.
// If the argument was not returned from MakeVariable, -1 is returned.
func GetElementErrorLevel(err error) int {
	if err, ok := err.(elemErr); ok {
		return err.level
	}
	return -1
}

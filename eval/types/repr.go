package types

import (
	"fmt"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// NoPretty can be passed to Repr to suppress pretty-printing.
const NoPretty = util.MinInt

// Reprer wraps the Repr method.
type Reprer interface {
	// Repr returns a string that represents a Value. The string either be a
	// literal of that Value that is preferably deep-equal to it (like `[a b c]`
	// for a list), or a string enclosed in "<>" containing the kind and
	// identity of the Value(like `<fn 0xdeadcafe>`).
	//
	// If indent is at least 0, it should be pretty-printed with the current
	// indentation level of indent; the indent of the first line has already
	// been written and shall not be written in Repr. The returned string
	// should never contain a trailing newline.
	Repr(indent int) string
}

// Repr returns the representation for a value, a string that is preferably (but
// not necessarily) an Elvish expression that evaluates to the argument. If
// indent >= 0, the representation is pretty-printed. It is implemented for the
// builtin types bool and string, and types satisfying the Reprer interface. For
// other types, it uses fmt.Sprint with the format "<unknown %v>".
func Repr(v interface{}, indent int) string {
	switch v := v.(type) {
	case bool:
		if v {
			return "$true"
		}
		return "$false"
	case string:
		return parse.Quote(v)
	case Reprer:
		return v.Repr(indent)
	default:
		return fmt.Sprintf("<unknown %v>", v)
	}
}
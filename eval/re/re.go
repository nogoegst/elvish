// Package re implements the re: module for using regular expressions.
package re

import (
	"fmt"
	"regexp"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/vector"
)

var Ns = eval.NewNs().AddBuiltinFns("re:", fns)

var fns = map[string]interface{}{
	"quote":   regexp.QuoteMeta,
	"match":   match,
	"find":    find,
	"replace": replace,
	"split":   split,
}

func match(opts eval.Options, argPattern, source string) bool {
	var optPOSIX bool
	opts.Scan(eval.OptToScan{"posix", &optPOSIX, false})

	pattern := makePattern(argPattern, optPOSIX, false)
	return pattern.MatchString(source)
}

func find(fm *eval.Frame, opts eval.Options, argPattern, source string) {
	out := fm.OutputChan()
	var (
		optPOSIX   bool
		optLongest bool
		optMax     int
	)
	opts.Scan(
		eval.OptToScan{"posix", &optPOSIX, false},
		eval.OptToScan{"longest", &optLongest, false},
		eval.OptToScan{"max", &optMax, "-1"})

	pattern := makePattern(argPattern, optPOSIX, optLongest)
	matches := pattern.FindAllSubmatchIndex([]byte(source), optMax)

	for _, match := range matches {
		start, end := match[0], match[1]
		groups := vector.Empty
		for i := 0; i < len(match); i += 2 {
			start, end := match[i], match[i+1]
			text := ""
			// FindAllSubmatchIndex may return negative indicies to indicate
			// that the pattern didn't appear in the text.
			if start >= 0 && end >= 0 {
				text = source[start:end]
			}
			groups = groups.Cons(newSubmatch(text, start, end))
		}
		out <- newMatch(source[start:end], start, end, groups)
	}
}

func replace(fm *eval.Frame, opts eval.Options,
	argPattern string, argRepl interface{}, source string) string {

	var (
		optPOSIX   bool
		optLongest bool
		optLiteral bool
	)
	opts.Scan(
		eval.OptToScan{"posix", &optPOSIX, false},
		eval.OptToScan{"longest", &optLongest, false},
		eval.OptToScan{"literal", &optLiteral, false})

	pattern := makePattern(argPattern, optPOSIX, optLongest)

	if optLiteral {
		repl, ok := argRepl.(string)
		if !ok {
			throwf("replacement must be string when literal is set, got %s",
				types.Kind(argRepl))
		}
		return pattern.ReplaceAllLiteralString(source, repl)
	} else {
		switch repl := argRepl.(type) {
		case string:
			return pattern.ReplaceAllString(source, repl)
		case eval.Callable:
			replFunc := func(s string) string {
				values, err := fm.PCaptureOutput(repl, []interface{}{s}, eval.NoOpts)
				maybeThrow(err)
				if len(values) != 1 {
					throwf("replacement function must output exactly one value, got %d", len(values))
				}
				output, ok := values[0].(string)
				if !ok {
					throwf("replacement function must output one string, got %s",
						types.Kind(values[0]))
				}
				return output
			}
			return pattern.ReplaceAllStringFunc(source, replFunc)
		default:
			throwf("replacement must be string or function, got %s",
				types.Kind(argRepl))
			panic("unreachable")
		}
	}
}

func split(fm *eval.Frame, opts eval.Options, argPattern, source string) {
	out := fm.OutputChan()
	var (
		optPOSIX   bool
		optLongest bool
		optMax     int
	)
	opts.Scan(
		eval.OptToScan{"posix", &optPOSIX, false},
		eval.OptToScan{"longest", &optLongest, false},
		eval.OptToScan{"max", &optMax, "-1"})

	pattern := makePattern(argPattern, optPOSIX, optLongest)

	pieces := pattern.Split(source, optMax)
	for _, piece := range pieces {
		out <- piece
	}
}

func makePattern(argPattern string, optPOSIX, optLongest bool) *regexp.Regexp {
	var (
		pattern *regexp.Regexp
		err     error
	)
	if optPOSIX {
		pattern, err = regexp.CompilePOSIX(string(argPattern))
	} else {
		pattern, err = regexp.Compile(string(argPattern))
	}
	maybeThrow(err)
	if optLongest {
		pattern.Longest()
	}
	return pattern
}

func throwf(format string, args ...interface{}) {
	util.Throw(fmt.Errorf(format, args...))
}

func maybeThrow(err error) {
	if err != nil {
		util.Throw(err)
	}
}

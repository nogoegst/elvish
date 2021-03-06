package eval

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/elves/elvish/eval/types"
)

// Input and output.

func init() {
	addBuiltinFns(map[string]interface{}{
		// Value output
		"put": put,

		// Bytes output
		"print":  print,
		"echo":   echo,
		"pprint": pprint,
		"repr":   repr,

		// Bytes to value
		"slurp":      slurp,
		"from-lines": fromLines,
		"from-json":  fromJSON,

		// Value to bytes
		"to-lines": toLines,
		"to-json":  toJSON,

		// File and pipe
		"fopen":   fopen,
		"fclose":  fclose,
		"pipe":    pipe,
		"prclose": prclose,
		"pwclose": pwclose,
	})
}

func put(fm *Frame, args ...interface{}) {
	out := fm.ports[1].Chan
	for _, a := range args {
		out <- a
	}
}

func print(fm *Frame, opts Options, args ...interface{}) {
	var sepv string
	opts.Scan(OptToScan{"sep", &sepv, " "})

	out := fm.ports[1].File
	sep := sepv
	for i, arg := range args {
		if i > 0 {
			out.WriteString(sep)
		}
		out.WriteString(types.ToString(arg))
	}
}

func echo(fm *Frame, opts Options, args ...interface{}) {
	print(fm, opts, args...)
	fm.ports[1].File.WriteString("\n")
}

func pprint(fm *Frame, args ...interface{}) {
	out := fm.ports[1].File
	for _, arg := range args {
		out.WriteString(types.Repr(arg, 0))
		out.WriteString("\n")
	}
}

func repr(fm *Frame, args ...interface{}) {
	out := fm.ports[1].File
	for i, arg := range args {
		if i > 0 {
			out.WriteString(" ")
		}
		out.WriteString(types.Repr(arg, types.NoPretty))
	}
	out.WriteString("\n")
}

func slurp(fm *Frame) (string, error) {
	b, err := ioutil.ReadAll(fm.ports[0].File)
	return string(b), err
}

func fromLines(fm *Frame) {
	linesToChan(fm.ports[0].File, fm.ports[1].Chan)
}

// fromJSON parses a stream of JSON data into Value's.
func fromJSON(fm *Frame) error {
	in := fm.ports[0].File
	out := fm.ports[1].Chan

	dec := json.NewDecoder(in)
	var v interface{}
	for {
		err := dec.Decode(&v)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		out <- FromJSONInterface(v)
	}
	return nil
}

func toLines(fm *Frame, inputs Inputs) {
	out := fm.ports[1].File

	inputs(func(v interface{}) {
		fmt.Fprintln(out, types.ToString(v))
	})
}

// toJSON converts a stream of Value's to JSON data.
func toJSON(fm *Frame, inputs Inputs) {
	out := fm.ports[1].File

	enc := json.NewEncoder(out)
	inputs(func(v interface{}) {
		err := enc.Encode(v)
		maybeThrow(err)
	})
}

func fopen(fm *Frame, name string) (types.File, error) {
	// TODO support opening files for writing etc as well.
	f, err := os.Open(name)
	return types.File{f}, err
}

func fclose(f types.File) error {
	return f.Inner.Close()
}

func pipe() (types.Pipe, error) {
	r, w, err := os.Pipe()
	return types.Pipe{r, w}, err
}

func prclose(p types.Pipe) error {
	return p.ReadEnd.Close()
}

func pwclose(p types.Pipe) error {
	return p.WriteEnd.Close()
}

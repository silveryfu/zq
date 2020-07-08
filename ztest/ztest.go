// Package ztest runs formulaic tests ("ztests") that can be (1) run in-process
// with the compiled-ini zq code base, (2) run as a sub-process using the zq
// executable build artifact, or (3) run as a bash script running a sequence
// of arbitrary shell commands invoking any of the build artifacts.  The
// first two cases comprise the "ZQL test style" and the last case
// comprises the "script test style".  Case (1) is easier to debug by
// simply running "go test" compared replicating the test using "go run".
// Script-style tests don't have this convenience.
//
// In the ZQL style, ztest runs a ZQL query on an input and checks
// for an expected output.
//
// A ZQL-style test is defined in a YAML file.
//
//    zql: count()
//
//    input: |
//      #0:record[i:int64]
//      0:[1;]
//      0:[2;]
//
//    output: |
//      #0:record[count:uint64]
//      0:[2;]
//
// Input format is detected automatically and can be anything recognized by
// "zq -i auto" (including optional gzip compression).  Output format defaults
// to tzng but can be set to anything accepted by "zq -f".
//
//    zql: count()
//
//    input: |
//      #0:record[i:int64]
//      0:[1;]
//      0:[2;]
//
//    output-format: table
//
//    output: |
//      COUNT
//      2
//
// Alternatively, tests can be configured to run as shell scripts.
// In this style of test, arbitrary bash scripts can run chaining together
// any of zq/cmd tools in addition to zq.  Here, the yaml sets up a collection
// of input files and stdin, the script runs, and the test driver compares expected
// output files, stdout, and stderr with data in the yaml spec.  In this case,
// instead of specifying, "zql", "input", "output", you specify the yaml arrays
// "inputs" and "outputs" --- where each array element defines a file, stdin,
// stdout, or stderr --- and a "script" that specifies a multi-line yaml string
// defining the script, e.g.,
//
// inputs:
//    - name: in1.tzng
//      data: |
//         #0:record[i:int64]
//         0:[1;]
//    - name: stdin
//      data: |
//         #0:record[i:int64]
//         0:[2;]
// script: |
//    zq -o out.tzng in1.tzng -
//    zq -o count.tzng "count()" out.tzng
// outputs:
//    - name: out.tzng
//      data: |
//         #0:record[i:int64]
//         0:[1;]
//         0:[2;]
//    - name: count.tzng
//      data: |
//         #0:record[count:uint64]
//         0:[2;]
//
// Each input and output has a name.  For inputs, a file (source),
// inlined data (data), or hexadecimal data (hex) may be specified.
// If no data is specified, then a file of the same name as the
// name field is looked for in the same directory as the yaml file.
// The source spec is a file path relative to the directory of the
// yaml file.  For outputs, expected output is defined in the same
// fashion as the inputs though you can also specify a "regexp" string
// instead of expected data.  If an output is named "stdout" or "stderr"
// then the actual output is taken from the stdout or stderr of the
// the shell script.
//
// Ztest YAML files for a package should reside in a subdirectory named
// testdata/ztest.
//
//     pkg/
//       pkg.go
//       pkg_test.go
//       testdata/
//         ztest/
//           test-1.yaml
//           test-2.yaml
//           ...
//
// Name YAML files descriptively since each ztest runs as a subtest
// named for the file that defines it.
//
// pkg_test.go should contain a Go test named TestZTest that calls Run.
//
//     func TestZTest(t *testing.T) { ztest.Run(t, "testdata/ztest") }
//
// If the ZTEST_BINDIR environment variable is unset or empty and the test
// is not a script test, Run runs ztests in the current process and skips
// the script tests.  Otherwise, Run runs each ztest in a separate process
// using the zq executable in the directory specified by ZTEST_BINDIR.
package ztest

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/brimsec/zq/driver"
	"github.com/brimsec/zq/emitter"
	"github.com/brimsec/zq/zbuf"
	"github.com/brimsec/zq/zio"
	"github.com/brimsec/zq/zio/detector"
	"github.com/brimsec/zq/zng/resolver"
	"github.com/brimsec/zq/zql"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Run runs the ztests in the directory named dirname.  For each file f.yaml in
// the directory, Run calls FromYAMLFile to load a ztest and then runs it in
// subtest named f.  bindir is a path to the executables that the script-mode
// tests will run.
func Run(t *testing.T, dirname string) {
	bindir := os.Getenv("ZTEST_BINDIR")
	if bindir != "" {
		if out, _, err := runzq(bindir, "help", "", ""); err != nil {
			if out != "" {
				out = fmt.Sprintf(" with output %q", out)
			}
			t.Fatalf("failed to exec zq in dir $ZTEST_BINDIR %s: %s%s", bindir, err, out)
		}
	}
	fileinfos, err := ioutil.ReadDir(dirname)
	if err != nil {
		t.Fatal(err)
	}
	for _, fi := range fileinfos {
		filename := fi.Name()
		const dotyaml = ".yaml"
		if !strings.HasSuffix(filename, dotyaml) {
			continue
		}
		testname := strings.TrimSuffix(filename, dotyaml)
		t.Run(testname, func(t *testing.T) {
			t.Parallel()
			// An absolute path in errors makes the offending file easier to find.
			filename, err := filepath.Abs(filepath.Join(dirname, filename))
			if err != nil {
				t.Fatal(err)
			}
			zt, err := FromYAMLFile(filename)
			if err != nil {
				t.Fatalf("%s: %s", filename, err)
			}
			zt.Run(t, testname, bindir, dirname, filename)
		})
	}
}

type File struct {
	// Name is the name of the file with respect to the directoy in which
	// the test script runs.  For inputs, if no data source is specified,
	// then name is also the name of a data file in the diectory containing
	// the yaml test file, which is copied to the test script directory.
	// Name can also be stdio (for inputs) or stdout or stderr (for outputs).
	Name string `yaml:"name"`
	// Data, Hex, and Source represents the different ways file data can
	// be defined for this file.  Data is a string turned into the contents
	// of the file, Hex is hex decoded, and Source is a string representing
	// the pathname of a file the repo that is read to comprise the data.
	Data   *string `yaml:"data,omitempty"`
	Hex    string  `yaml:"hex,omitempty"`
	Source string  `yaml:"source,omitempty"`
	// Re is a regular expression describing the contents of the file,
	// which is only applicable to output files.
	Re string `yaml:"regexp,omitempty"`
}

func (f *File) check() error {
	cnt := 0
	if f.Data != nil {
		cnt++
	}
	if f.Hex != "" {
		cnt++
	}
	if f.Source != "" {
		cnt++
	}
	if cnt > 1 {
		return fmt.Errorf("%s: must at most one of data, hex, or source", f.Name)
	}
	return nil
}

func (f *File) load(dir string) ([]byte, *regexp.Regexp, error) {
	if f.Data != nil {
		return []byte(*f.Data), nil, nil
	}
	if f.Hex != "" {
		s, err := decodeHex(f.Hex)
		return []byte(s), nil, err
	}
	if f.Source != "" {
		b, err := ioutil.ReadFile(filepath.Join(dir, f.Source))
		return b, nil, err
	}
	if f.Re != "" {
		re, err := regexp.Compile(f.Re)
		return nil, re, err
	}
	b, err := ioutil.ReadFile(filepath.Join(dir, f.Name))
	if err == nil {
		return b, nil, nil
	}
	if os.IsNotExist(err) {
		err = fmt.Errorf("%s: no data source", f.Name)
	}
	return nil, nil, err
}

// ZTest defines a ztest.
type ZTest struct {
	ZQL          string `yaml:"zql,omitempty"`
	Input        Inputs `yaml:"input,omitempty"`
	OutputFormat string `yaml:"output-format,omitempty"`
	Output       string `yaml:"output,omitempty"`
	OutputHex    string `yaml:"outputHex,omitempty"`
	OutputFlags  string `yaml:"output-flags,omitempty"`
	ErrorRE      string `yaml:"errorRE"`
	errRegex     *regexp.Regexp
	Warnings     string `yaml:"warnings",omitempty"`
	// shell mode params
	Script  string `yaml:"script,omitempty"`
	Inputs  []File `yaml:"inputs,omitempty"`
	Outputs []File `yaml:"outputs,omitempty"`
}

func (z *ZTest) check() error {
	if z.ZQL != "" {
		if z.Input == nil {
			return errors.New("input field missing in a zq test")
		}
	} else if z.Script != "" {
		if z.Outputs == nil {
			return errors.New("outputs field missing in a sh test")
		}
		for _, f := range z.Inputs {
			if err := f.check(); err != nil {
				return err
			}
		}
		for _, f := range z.Outputs {
			if err := f.check(); err != nil {
				return err
			}
		}
	} else {
		return errors.New("either a zql field or script field must be present")
	}
	return nil
}

// Inputs is an array of strings. Its only purpose is to support parsing of
// both single string and array yaml values for the field ZTest.Input.
type Inputs []string

func (i *Inputs) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.SequenceNode {
		var inputs []string
		err := value.Decode(&inputs)
		*i = inputs
		return err
	}
	var input string
	if err := value.Decode(&input); err != nil {
		return err
	}
	*i = append(*i, input)
	return nil
}

// Try to decode a yaml-friendly way of representing binary data in hex:
// each line is either a comment explaining the contents (denoted with
// a leading # character), or a sequence of hex digits.
func decodeHex(in string) (string, error) {
	var raw string
	for _, line := range strings.Split(in, "\n") {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		raw += strings.ReplaceAll(line, " ", "")
	}
	out := make([]byte, hex.DecodedLen(len(raw)))
	_, err := hex.Decode(out, []byte(raw))
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func encodeHex(in string) string {
	var buf bytes.Buffer
	dumper := hex.Dumper(&buf)
	dumper.Write([]byte(in))
	return buf.String()
}

func (z *ZTest) getOutput() (string, error) {
	outlen := len(z.Output)
	hexlen := len(z.OutputHex)
	if outlen > 0 && hexlen > 0 {
		return "", errors.New("Cannot specify both output and outputHex")
	}
	if outlen == 0 && hexlen == 0 {
		return "", nil
	}
	if outlen > 0 {
		return z.Output, nil
	}
	return decodeHex(z.OutputHex)
}

// FromYAMLFile loads a ZTest from the YAML file named filename.
func FromYAMLFile(filename string) (*ZTest, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	d := yaml.NewDecoder(bytes.NewReader(buf))
	d.KnownFields(true)
	var z ZTest
	if err := d.Decode(&z); err != nil {
		return nil, err
	}
	var v interface{}
	if d.Decode(&v) != io.EOF {
		return nil, errors.New("found multiple YAML documents or garbage after first document")
	}
	if z.OutputFormat == "" {
		z.OutputFormat = "tzng"
	}
	if z.ErrorRE != "" {
		z.errRegex, err = regexp.Compile(z.ErrorRE)
		if err != nil {
			return nil, err
		}
	}
	return &z, nil
}

func diffErr(expected, actual string) error {
	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(expected),
		FromFile: "expected",
		B:        difflib.SplitLines(actual),
		ToFile:   "actual",
		Context:  5,
	})
	return fmt.Errorf("expected and actual outputs differ:\n%s", diff)
}

func (z *ZTest) Run(t *testing.T, testname, bindir, dirname, filename string) {
	if err := z.check(); err != nil {
		t.Fatalf("%s: bad yaml format: %s", filename, err)
	}
	if z.Script != "" {
		if runtime.GOOS == "windows" {
			// XXX skip in windows until we figure out the best
			// way to support script-driven tests across
			// environments
			t.Skip("skipping script test on Windows")
		}
		if bindir == "" {
			t.Skip("skipping script test on in-process run")
		}
		adir, _ := filepath.Abs(dirname)
		err := runsh(testname, bindir, adir, z)
		if err != nil {
			t.Fatalf("%s: %s", filename, err)
		}
		return
	}
	out, errout, err := runzq(bindir, z.ZQL, z.OutputFormat, z.OutputFlags, z.Input...)
	if err != nil {
		if z.errRegex != nil {
			if !z.errRegex.Match([]byte(errout)) {
				t.Fatalf("%s: error doesn't match expected error regex: %s %s", filename, z.ErrorRE, errout)
			}
		} else {
			if out != "" {
				out = "\noutput:\n" + out
			}
			t.Fatalf("%s: %s%s", filename, err, out)
		}
	} else if z.errRegex != nil {
		t.Fatalf("%s: no error when expecting error regex: %s", filename, z.ErrorRE)
	}
	expectedOut, oerr := z.getOutput()
	require.NoError(t, oerr)
	if out != expectedOut {
		a := expectedOut
		b := out

		if !utf8.ValidString(a) {
			a = encodeHex(a)
			b = encodeHex(b)
		}

		err := diffErr(a, b)
		t.Fatalf("%s: expected and actual outputs differ:\n%s", filename, err)
	}
	if err == nil && errout != z.Warnings {
		err := diffErr(z.Warnings, errout)
		t.Fatalf("%s: expected and actual warnings differ:\n%s", filename, err)
	}
}

func checkPatterns(patterns map[string]*regexp.Regexp, dir *Dir, stdout, stderr string) error {
	for name, re := range patterns {
		var body []byte
		switch name {
		case "stdout":
			body = []byte(stdout)
		case "stderr":
			body = []byte(stderr)
		default:
			var err error
			body, err = dir.Read(name)
			if err != nil {
				return fmt.Errorf("%s: %s", name, err)
			}
		}
		if !re.Match(body) {
			return fmt.Errorf("regex mismatch: '%s' does not match '%s'", re, string(body))
		}
	}
	return nil
}

func checkData(files map[string][]byte, dir *Dir, stdout, stderr string) error {
	for name, expected := range files {
		var actual []byte
		switch name {
		case "stdout":
			actual = []byte(stdout)
		case "stderr":
			actual = []byte(stderr)
		default:
			var err error
			actual, err = dir.Read(name)
			if err != nil {
				return fmt.Errorf("%s: %s", name, err)
			}
		}
		if !bytes.Equal(expected, actual) {
			return diffErr(string(expected), string(actual))
		}
	}
	return nil
}

func runsh(testname, bindir, dirname string, zt *ZTest) error {
	dir, err := NewDir(testname)
	if err != nil {
		return err
	}
	var stdin io.Reader
	defer dir.RemoveAll()
	for _, f := range zt.Inputs {
		b, re, err := f.load(dirname)
		if err != nil {
			return err
		}
		if f.Name == "stdin" {
			stdin = bytes.NewReader(b)
			continue
		}
		if re != nil {
			return fmt.Errorf("%s: cannot use a regexp pattern in an input", f.Name)
		}
		if err := dir.Write(f.Name, b); err != nil {
			return err
		}
	}
	expectedData := make(map[string][]byte)
	expectedPattern := make(map[string]*regexp.Regexp)
	for _, f := range zt.Outputs {
		b, re, err := f.load(dirname)
		if err != nil {
			return err
		}
		if b != nil {
			expectedData[f.Name] = b
		}
		if re != nil {
			expectedPattern[f.Name] = re
		}
	}
	stdout, stderr, err := RunShell(dir, bindir, zt.Script, stdin)
	if err != nil {
		// XXX If the err is an exit error, we ignore it and rely on
		// tests that check stderr etc.  We could pull out the exit
		// status and test on this if we added a field for this to
		// the ZTest struct.  I don't think it makes sense to comingle
		// this condition with the stderr checks as in the other
		// testing code path below.
		if _, ok := err.(*exec.ExitError); !ok {
			// Not an exit error from the test shell so there was
			// a problem execing and runnning the shell command...
			return err
		}
	}
	err = checkPatterns(expectedPattern, dir, stdout, stderr)
	if err != nil {
		return err
	}
	return checkData(expectedData, dir, stdout, stderr)
}

// runzq runs the query in ZQL over inputs and returns the output formatted
// according to outputFormat. inputs may be in any format recognized by "zq -i
// auto" and maybe be gzip-compressed.  outputFormat may be any string accepted
// by "zq -f".  If bindir is empty, the query runs in the current process.
// If bindir is not empty, it specifies a zq path that will be used to run
// the query.
func runzq(bindir, ZQL, outputFormat, outputFlags string, inputs ...string) (out string, warnOrError string, err error) {
	var outbuf bytes.Buffer
	var errbuf bytes.Buffer
	if bindir != "" {
		zq := filepath.Join(bindir, "zq")
		tmpdir, files, err := tmpInputFiles(inputs)
		if err != nil {
			return "", "", err
		}
		defer os.RemoveAll(tmpdir)
		cmd := exec.Command(zq, "-f", outputFormat)
		if len(outputFlags) > 0 {
			flags := strings.Split(outputFlags, " ")
			cmd.Args = append(cmd.Args, flags...)
		}
		cmd.Args = append(cmd.Args, ZQL)
		cmd.Args = append(cmd.Args, files...)
		cmd.Stdout = &outbuf
		cmd.Stderr = &errbuf
		err = cmd.Run()
		// If there was an error, errbuf could potentially hold both warnings
		// and error messages, but that's not currently an issue with existing
		// tests.
		return string(outbuf.Bytes()), string(errbuf.Bytes()), err
	}
	proc, err := zql.ParseProc(ZQL)
	if err != nil {
		return "", "", err
	}
	zctx := resolver.NewContext()
	zr, err := loadInputs(inputs, zctx)
	if err != nil {
		return "", err.Error(), err
	}
	defer zr.Close()
	if outputFormat == "types" {
		outputFormat = "null"
		zctx.SetLogger(&emitter.TypeLogger{WriteCloser: &nopCloser{&outbuf}})
	}
	muxOutput, err := driver.Compile(context.Background(), proc, zr, driver.Config{TypeContext: zctx})
	if err != nil {
		return "", "", err
	}
	var zflags zio.WriterFlags
	var flags flag.FlagSet
	zflags.SetFlags(&flags)
	err = flags.Parse(strings.Split(outputFlags, " "))
	if err != nil {
		return "", "", err
	}
	zflags.Format = outputFormat
	zw := detector.LookupWriter(&nopCloser{&outbuf}, &zflags)
	if zw == nil {
		return "", "", fmt.Errorf("%s: unknown output format", outputFormat)
	}
	d := driver.NewCLI(zw)
	d.SetWarningsWriter(&errbuf)
	err = driver.Run(muxOutput, d, nil)
	if err2 := zw.Flush(); err == nil {
		err = err2
	}
	if err != nil {
		return string(outbuf.Bytes()), err.Error(), err
	}
	return string(outbuf.Bytes()), string(errbuf.Bytes()), nil
}

func loadInputs(inputs []string, zctx *resolver.Context) (*zbuf.Combiner, error) {
	var readers []zbuf.Reader
	for _, input := range inputs {
		zr, err := detector.NewReader(detector.GzipReader(strings.NewReader(input)), zctx)
		if err != nil {
			return nil, err
		}
		readers = append(readers, zr)
	}
	return zbuf.NewCombiner(readers, zbuf.CmpTimeForward), nil
}

func tmpInputFiles(inputs []string) (string, []string, error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", nil, err
	}
	var files []string
	for i, input := range inputs {
		name := fmt.Sprintf("input%d", i+1)
		file := filepath.Join(dir, name)
		err := ioutil.WriteFile(file, []byte(input), 0644)
		if err != nil {
			os.RemoveAll(dir)
			return "", nil, err
		}
		files = append(files, file)
	}
	return dir, files, nil
}

type nopCloser struct{ io.Writer }

func (*nopCloser) Close() error { return nil }

package create

import (
	"errors"
	"flag"

	"github.com/brimsec/zq/cli/inputflags"
	"github.com/brimsec/zq/cli/outputflags"
	"github.com/brimsec/zq/cmd/zst/root"
	"github.com/brimsec/zq/zbuf"
	"github.com/brimsec/zq/zng/resolver"
	"github.com/mccanne/charm"
)

var Create = &charm.Spec{
	Name:  "create",
	Usage: "create [-coltresh thresh] [-skewthresh thesh] -o file files...",
	Short: "create a zst columnar object from a zng file or stream",
	Long: `
The create command generates a columnar zst object from a zng input stream,
which may be stdin or one or more zng storage objects (local files or s3 objects).
The output can be a local file or an s3 URI.

The -colthresh flag specifies the byte threshold (in MiB) at which chunks
of column data are written to disk.

The -skewthresh flag specifies a rough byte threshold (in MiB) that controls
how much column data is collectively buffered in memory before being entirely
flushed to disk.  This parameter controls the amount of buffering "skew" required
keep rows in alignment so that a reader should not have to use more than
this (approximate) memory footprint.

Unlike parquet, zst column data may be laid out any way a client so chooses
and is not constrained to the "row group" concept.  Thus, care should be
taken here to control the amount of row skew that can arise.`,
	New: newCommand,
}

func init() {
	root.Zst.Add(Create)
}

type Command struct {
	*root.Command
	outputFlags outputflags.Flags
	inputFlags  inputflags.Flags
}

func MibToBytes(mib float64) int {
	return int(mib * 1024 * 1024)
}

func newCommand(parent charm.Command, f *flag.FlagSet) (charm.Command, error) {
	c := &Command{
		Command: parent.(*root.Command),
	}
	c.inputFlags.SetFlags(f)
	c.outputFlags.SetFlagsWithFormat(f, "zst")
	return c, nil
}

func (c *Command) Run(args []string) error {
	defer c.Cleanup()
	if err := c.Init(&c.inputFlags, &c.outputFlags); err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("must specify one or more input files")
	}
	zctx := resolver.NewContext()
	readers, err := c.inputFlags.Open(zctx, args, true)
	if err != nil {
		return err
	}
	reader := zbuf.NewCombiner(readers, zbuf.CmpTimeForward)
	defer reader.Close()
	writer, err := c.outputFlags.Open()
	if err != nil {
		return err
	}
	if err := zbuf.Copy(writer, reader); err != nil {
		return err
	}
	return writer.Close()
}

package version

import (
	"errors"
	"flag"
	"fmt"

	"github.com/brimsec/zq/cmd/zapi/cmd"
	"github.com/mccanne/charm"
)

var Version = &charm.Spec{
	Name:  "version",
	Usage: "version",
	Short: "show version of connected zqd",
	Long: `
The version command displays the version string of the connected zqd.
Use -version to show the version string of the zapi tool.`,
	New: New,
}

func init() {
	cmd.CLI.Add(Version)
}

type Command struct {
	*cmd.Command
}

func New(parent charm.Command, flags *flag.FlagSet) (charm.Command, error) {
	return &Command{Command: parent.(*cmd.Command)}, nil
}

// Run lists all spaces in the current zqd host or if a parameter
// is provided (in glob style) lists the info about that space.
func (c *Command) Run(args []string) error {
	client := c.Client()
	if len(args) > 0 {
		return errors.New("version command takes no arguemtns")
	}
	version, err := client.Version(c.Context())
	if err != nil {
		return err
	}
	fmt.Println(version)
	return nil
}

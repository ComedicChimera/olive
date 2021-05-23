package olive

import (
	"fmt"
	"log"
	"os"
)

// This file outlines the user-facing API of Olive.
// -----------------------------------------------------------------------------

// Command represents a keyword which determines the state of the parser and how
// it should continue to parse: it is a directive or subdirective to the
// application.  (eg. `go build` has a command of `go` and a subcommand of
// `build`).
type Command struct {
	// Name is the name of the command
	Name string

	// Description is a descriptive string for the command
	Description string

	// Requires subcommand indicates if this command expects a subcommand or can
	// be satisfied without one
	RequiresSubcommand bool

	// All valid subcommands of this command organized by name.  The flag
	// indicates whether or not a subcommand must be provided.
	subcommands map[string]*Command

	// Flags and named arguments organized by their full name
	flags map[string]*Flag
	args  map[string]Argument

	// Flags and named arguments organized by their short name for quick access
	// during parsing
	flagsByShortName map[string]*Flag
	argsByShortName  map[string]Argument

	// There can only be one primary argument per command
	primaryArg *PrimaryArgument
}

// ArgParseResult is the result produced by the argument parser representing the
// inputted arguments if parsing succeeded.
type ArgParseResult struct {
	flags map[string]struct{}

	Arguments map[string]interface{}

	subcommandName string
	subcommandRes  *ArgParseResult

	primaryArg string
}

// -----------------------------------------------------------------------------

// NewCLI creates a new CLI (initial command) to be customized by the user
func NewCLI(name, desc string, helpEnabled bool) *Command {
	return newCommand(name, desc, helpEnabled)
}

// ParseArgs parses the slice of arguments provided against a customized CLI. It
// returns an ArgParseResult representing the accumulated result of parsing and
// an error which will be `nil` if no error occured
func ParseArgs(cli *Command, args []string) (*ArgParseResult, error) {
	ap := &argParser{initialCommand: cli}

	// trim off the first argument which is conventionally the application name
	return ap.parse(args[1:])
}

// -----------------------------------------------------------------------------

// AddSubcommand adds a subcommand to the command
func (c *Command) AddSubcommand(name, desc string, helpEnabled bool) *Command {
	if c.primaryArg != nil {
		log.Fatalf("command `%s` cannot both take a primary argument and have subcommands", c.Name)
	}

	if _, ok := c.subcommands[name]; ok {
		log.Fatalf("multiple subcommands named `%s`", name)
	}

	subc := newCommand(name, desc, helpEnabled)

	c.subcommands[name] = subc
	return subc
}

// AddPrimaryArg adds a primary argument to the command
func (c *Command) AddPrimaryArg(name, desc string, required bool) *PrimaryArgument {
	if len(c.subcommands) > 0 {
		log.Fatalf("command `%s` cannot both take a primary argument and have subcommands", c.Name)
	}

	c.primaryArg = &PrimaryArgument{name: name, desc: desc, required: required}
	return c.primaryArg
}

// AddFlag adds a flag to the command
func (c *Command) AddFlag(name, shortName, desc string) *Flag {
	if _, ok := c.flags[name]; ok {
		log.Fatalf("multiple flags named `%s`\n", name)
	}

	if _, ok := c.flagsByShortName[shortName]; ok {
		log.Fatalf("multiple flags with short name `%s`\n", shortName)
	}

	f := &Flag{
		name:      name,
		shortName: shortName,
		desc:      desc,
	}

	c.flags[name] = f
	c.flagsByShortName[shortName] = f

	return f
}

// AddIntArg adds a named integer argument
func (c *Command) AddIntArg(name, shortName, desc string, required bool) *IntArgument {
	ia := &IntArgument{
		argumentBase: argumentBase{
			name:      name,
			shortName: shortName,
			desc:      desc,
			required:  required,
		},
	}

	c.addArg(ia)
	return ia
}

// AddFloatArg adds a named float argument
func (c *Command) AddFloatArg(name, shortName, desc string, required bool) *FloatArgument {
	fa := &FloatArgument{
		argumentBase: argumentBase{
			name:      name,
			shortName: shortName,
			desc:      desc,
			required:  required,
		},
	}

	c.addArg(fa)
	return fa
}

// AddStringArg adds a named string argument
func (c *Command) AddStringArg(name, shortName, desc string, required bool) *StringArgument {
	sa := &StringArgument{
		argumentBase: argumentBase{
			name:      name,
			shortName: shortName,
			desc:      desc,
			required:  required,
		},
	}

	c.addArg(sa)
	return sa
}

// AddSelectorArg adds a named selector argument
func (c *Command) AddSelectorArg(name, shortName, desc string, required bool, possibleValues []string) *SelectorArgument {
	pvals := make(map[string]struct{})
	for _, pval := range possibleValues {
		pvals[pval] = struct{}{}
	}

	sa := &SelectorArgument{
		argumentBase: argumentBase{
			name:      name,
			shortName: shortName,
			desc:      desc,
			required:  required,
		},
		possibleValues: pvals,
	}

	c.addArg(sa)
	return sa
}

// addArg adds an argument to a command
func (c *Command) addArg(arg Argument) {
	if _, ok := c.args[arg.Name()]; ok {
		log.Fatalf("multiple arguments named `%s`", arg.Name())
	}

	if _, ok := c.argsByShortName[arg.ShortName()]; ok {
		log.Fatalf("multiple arguments with short name `%s`", arg.ShortName())
	}

	c.args[arg.Name()] = arg
	c.argsByShortName[arg.ShortName()] = arg
}

// EnableHelp enables the help flag (`--help` or `-h`).
func (c *Command) EnableHelp() {
	if _, ok := c.args["help"]; !ok {
		flag := c.AddFlag("help", "h", "Get help")
		flag.action = func() {
			c.Help()
			os.Exit(0)
		}
	}
}

// DisableHelp disables the help flag (`--help` or `-h`).
func (c *Command) DisableHelp() {
	if _, ok := c.flags["help"]; ok {
		delete(c.flags, "help")
		delete(c.flagsByShortName, "h")
	}
}

// -----------------------------------------------------------------------------

// HasFlag checks if a flag has been set during argument parsing
func (apr *ArgParseResult) HasFlag(name string) bool {
	_, ok := apr.flags[name]
	return ok
}

// PrimaryArg gets the primary argument if one exists
func (apr *ArgParseResult) PrimaryArg() (string, bool) {
	return apr.primaryArg, apr.primaryArg != ""
}

// Subcommand gets the subcommand if one exists
func (apr *ArgParseResult) Subcommand() (string, *ArgParseResult, bool) {
	return apr.subcommandName, apr.subcommandRes, apr.subcommandRes != nil
}

// -----------------------------------------------------------------------------

// Help displays the help message for a given command
func (c *Command) Help() {
	fmt.Println(getHelpMessage(c))
}

// HelpMessage returns the stringified help message for a given command
func (c *Command) HelpMessage() string {
	return getHelpMessage(c)
}

package olive

import (
	"fmt"
	"strings"
)

// argParser is a state machine used to parse arguments
type argParser struct {
	// initialCommand is the command that represents the initial/global state of
	// the parser.  The entire CLI of an application can essentially be thought of
	// as one large command which a bunch of subcommands.
	initialCommand *Command

	// commandStack is a stack of active commands.  This facilitates the fact
	// that flags and named arguments that are valid for a base command are also
	// valid for all subcommands.
	commandStack []*Command

	// result is the accumulated result of parsing.  This data structure is
	// built up during the parsing process to represent the abstract structure
	// of the inputted arguments -- it is akin to an AST in language processing.
	result *ArgParseResult

	// semanticStack stores the components of the arg parse result as they are
	// constructed -- this allows for flags to be assigned to their
	// corresponding command at any depth within the parsing stack
	semanticStack []*ArgParseResult

	// allowSubcommands indicates whether or not a flag or argument has already
	// been encountered and therefore subcommands are no longer valid
	allowSubcommands bool
}

// parse runs the main parsing algorithm on a set of argument values
func (ap *argParser) parse(args []string) (*ArgParseResult, error) {
	ap.result = &ArgParseResult{
		flags:     make(map[string]struct{}),
		Arguments: make(map[string]interface{}),
	}
	ap.commandStack = []*Command{ap.initialCommand}
	ap.semanticStack = []*ArgParseResult{ap.result}
	ap.allowSubcommands = true

	for _, arg := range args {
		if err := ap.consume(arg); err != nil {
			return nil, err
		}
	}

	// by definition, the last value on the command stack can be the only
	// command that might be missing a subcommand -- so that is the only value
	// we check.  We know that if the last item on the command stack requires a
	// subcommand, then it is missing that command (otherwise, there would be a
	// next item).  We only check this field if there are subcommands to be
	// missing
	if len(ap.currCommand().subcommands) > 0 && ap.currCommand().RequiresSubcommand {
		return nil, fmt.Errorf("`%s` requires a subcommand", ap.currCommand().Name)
	}

	// set all the default values of any unsupplied arguments; go in reverse
	// order so most specific subcommand gets precedence
	for i := len(ap.commandStack) - 1; i > -1; i-- {
		for _, arg := range ap.commandStack[i].args {
			if val, ok := arg.GetDefaultValue(); ok {
				if _, ok := ap.semanticStack[i].Arguments[arg.Name()]; !ok {
					ap.semanticStack[i].Arguments[arg.Name()] = val
				}
			}
		}
	}

	return ap.result, nil
}

// consume processes a single argument token of input
func (ap *argParser) consume(arg string) error {
	if strings.HasPrefix(arg, "--") {
		ap.allowSubcommands = false

		// handle full-named arguments
		argName, argVal := ap.extractComponents(arg)

		if argVal == "" {
			// => flag
			for i := len(ap.commandStack) - 1; i > -1; i-- {
				if flag, ok := ap.commandStack[i].flags[argName]; ok {
					if err := ap.setFlag(i, flag); err != nil {
						return err
					} else {
						return nil
					}
				}
			}

			return fmt.Errorf("unknown flag: `%s`", argName)
		} else {
			// => argument
			for i := len(ap.commandStack) - 1; i > -1; i-- {
				if arg, ok := ap.commandStack[i].args[argName]; ok {
					if err := ap.setArg(i, arg, argVal); err != nil {
						return err
					} else {
						return nil
					}
				}
			}

			return fmt.Errorf("unknown argument: `%s`", argName)
		}
	} else if strings.HasPrefix(arg, "-") {
		ap.allowSubcommands = false

		// handle short-named arguments
		argName, argVal := ap.extractComponents(arg)

		if argVal == "" {
			// => flag
			for i := len(ap.commandStack) - 1; i > -1; i-- {
				if flag, ok := ap.commandStack[i].flagsByShortName[argName]; ok {
					if err := ap.setFlag(i, flag); err != nil {
						return err
					} else {
						return nil
					}
				}
			}

			return fmt.Errorf("unknown flag by short name: `%s`", argName)
		} else {
			// => argument
			for i := len(ap.commandStack) - 1; i > -1; i-- {
				if arg, ok := ap.commandStack[i].argsByShortName[argName]; ok {
					if err := ap.setArg(i, arg, argVal); err != nil {
						return err
					} else {
						return nil
					}
				}
			}

			return fmt.Errorf("unknown argument by short name: `%s`", argName)
		}
	} else if ap.currCommand().primaryArg != nil {
		ap.allowSubcommands = false

		// handle primary arguments
		if ap.currResult().primaryArg != "" {
			return fmt.Errorf("multiple primary arguments specified for command `%s`", ap.currCommand().Name)
		}

		ap.currResult().primaryArg = arg
	} else if ap.allowSubcommands {
		if subc, ok := ap.currCommand().subcommands[arg]; ok {
			// handle subcommands
			ap.commandStack = append(ap.commandStack, subc)

			newResult := &ArgParseResult{
				Arguments: make(map[string]interface{}),
				flags:     make(map[string]struct{}),
			}

			ap.currResult().subcommandRes = newResult
			ap.currResult().subcommandName = subc.Name
			ap.semanticStack = append(ap.semanticStack, newResult)
		} else {
			return fmt.Errorf("unknown subcommand: `%s`", arg)
		}
	} else {
		return fmt.Errorf("unknown subcommand: `%s`", arg)
	}

	return nil
}

// extractComponents converts an input string into its two parts: argument name
// and argument value.  If this input string is setting a flag, then the
// argument value returned is "".
func (ap *argParser) extractComponents(arg string) (string, string) {
	if strings.Contains(arg, "=") {
		argComponents := strings.Split(arg, "=")

		return strings.TrimLeft(argComponents[0], "-"), strings.Join(argComponents[1:], "=")
	} else {
		return strings.TrimLeft(arg, "-"), ""
	}
}

// setFlag attempts to set a flag in the parse result.  The input index is the
// result's position in the semantic stack.  This function returns an error if
// the flag is set multiple times.
func (ap *argParser) setFlag(ndx int, flag *Flag) error {
	if _, ok := ap.semanticStack[ndx].flags[flag.name]; ok {
		return fmt.Errorf("flag `%s` set multiple times", flag.name)
	}

	ap.semanticStack[ndx].flags[flag.name] = struct{}{}

	if flag.action != nil {
		flag.action()
	}

	return nil
}

// setArg attempts to set the value for an argument in the parse result.
// The input index is the result's position in the semantic stack.
func (ap *argParser) setArg(ndx int, arg Argument, value string) error {
	if _, ok := ap.semanticStack[ndx].Arguments[arg.Name()]; ok {
		return fmt.Errorf("argument `%s` set multiple times", arg.Name())
	}

	val, err := arg.checkValue(value)
	if err == nil {
		ap.semanticStack[ndx].Arguments[arg.Name()] = val
		return nil
	}

	return err
}

// currCommand returns the command on top of the command stack
func (ap *argParser) currCommand() *Command {
	return ap.commandStack[len(ap.commandStack)-1]
}

// currResult returns the result on top of the semantic stack
func (ap *argParser) currResult() *ArgParseResult {
	return ap.semanticStack[len(ap.semanticStack)-1]
}

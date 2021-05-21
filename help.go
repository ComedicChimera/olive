package olive

import (
	"fmt"
	"strings"

	"github.com/eidolon/wordwrap"
)

// helpBuilder is a type used to build help messages
type helpBuilder struct {
	c *Command
	b strings.Builder
	w wordwrap.WrapperFunc
}

// getHelpMessage generates a help message for a given command
func getHelpMessage(c *Command) string {
	hb := &helpBuilder{
		c: c,
		b: strings.Builder{},
		w: wordwrap.Wrapper(60, false),
	}

	return hb.buildMessage()
}

// -----------------------------------------------------------------------------

func (hb *helpBuilder) buildMessage() string {
	hb.b.WriteString(hb.w(hb.c.Description))
	hb.b.WriteString("\n\nUsage:\n\n")

	hb.buildUsageLine()

	if len(hb.c.subcommands) > 0 {
		hb.b.WriteString("\nCommands:\n\n")

		hb.buildSubcommandsList()
	}

	if hb.c.primaryArg != nil {
		hb.b.WriteString("\nPrimary Argument:\n\n")

		hb.b.WriteString(wordwrap.Indent(
			fmt.Sprintf("%s   %s", hb.c.primaryArg.name, hb.c.primaryArg.desc), "    ", false),
		)
	}

	if len(hb.c.args) > 0 {
		hb.b.WriteString("\nArguments:\n\n")

		hb.buildArgumentsList()
	}

	if len(hb.c.flags) > 0 {
		hb.b.WriteString("\nFlags:\n\n")

		hb.buildFlagsList()
	}

	return hb.b.String()
}

func (hb *helpBuilder) buildUsageLine() {
	ub := strings.Builder{}

	ub.WriteString(hb.c.Name + " ")

	if len(hb.c.subcommands) > 0 {
		ub.WriteString("<command> ")
	} else if hb.c.primaryArg != nil {
		ub.WriteString(fmt.Sprintf("[%s] ", hb.c.primaryArg.name))
	}

	for _, arg := range hb.c.args {
		var argValue string

		switch v := arg.(type) {
		case *IntArgument:
			argValue = "int"
		case *FloatArgument:
			argValue = "float"
		case *StringArgument:
			argValue = "string"
		case *SelectorArgument:
			vnamesB := strings.Builder{}
			for value := range v.possibleValues {
				vnamesB.WriteString(value)
				vnamesB.WriteRune('|')
			}

			argValue = vnamesB.String()[:vnamesB.Len()-1]
		}

		ub.WriteString(fmt.Sprintf("[%s|%s=<%s>] ", arg.ShortName(), arg.Name(), argValue))
	}

	for _, flag := range hb.c.flags {
		ub.WriteString(fmt.Sprintf("[-%s|--%s] ", flag.shortName, flag.name))
	}

	ub.WriteRune('\n')

	hb.b.WriteString(wordwrap.Indent(ub.String(), "    ", true))
}

func (hb *helpBuilder) buildSubcommandsList() {
	maxCmdNameColLength := 0
	for cmdName := range hb.c.subcommands {
		if len(cmdName) > maxCmdNameColLength {
			maxCmdNameColLength = len(cmdName)
		}
	}

	// 3 spaces to the right
	maxCmdNameColLength += 3

	// 4 spaces to the left
	wdesc := wordwrap.Wrapper(60-maxCmdNameColLength-4, false)

	for _, cmd := range hb.c.subcommands {
		hb.b.WriteString(wordwrap.Indent(
			wdesc(cmd.Description),
			"    "+cmd.Name+strings.Repeat(" ", maxCmdNameColLength-len(cmd.Name)),
			false,
		))

		hb.b.WriteRune('\n')
	}
}

func (hb *helpBuilder) buildArgumentsList() {
	maxArgNameColLength := 0
	for argName, arg := range hb.c.args {
		if len(argName)+len(arg.ShortName()) > maxArgNameColLength {
			maxArgNameColLength = len(argName) + len(arg.ShortName())
		}
	}

	// one comma, one space, 3 dashes
	maxArgNameColLength += 5

	// 4 spaces to the left
	wdesc := wordwrap.Wrapper(60-maxArgNameColLength-4, false)

	for _, arg := range hb.c.args {
		hb.b.WriteString(wordwrap.Indent(
			wdesc(arg.Description()),
			fmt.Sprintf(
				"    -%s, --%s%s   ",
				arg.ShortName(),
				arg.Name(),
				strings.Repeat(" ", maxArgNameColLength-len(arg.Name())-len(arg.ShortName())-5),
			),
			false,
		))

		hb.b.WriteRune('\n')
	}
}

func (hb *helpBuilder) buildFlagsList() {
	maxFlagNameColLength := 0
	for flagName, flag := range hb.c.flags {
		if len(flagName)+len(flag.shortName) > maxFlagNameColLength {
			maxFlagNameColLength = len(flagName) + len(flag.shortName)
		}
	}

	// one comma, one space, 3 dashes
	maxFlagNameColLength += 5

	// 4 spaces to the left
	wdesc := wordwrap.Wrapper(60-maxFlagNameColLength-4, false)

	for _, flag := range hb.c.flags {
		hb.b.WriteString(wordwrap.Indent(
			wdesc(flag.desc),
			fmt.Sprintf(
				"    -%s, --%s%s   ",
				flag.shortName,
				flag.name,
				strings.Repeat(" ", maxFlagNameColLength-len(flag.name)-len(flag.shortName)-5),
			),
			false,
		))

		hb.b.WriteRune('\n')
	}
}

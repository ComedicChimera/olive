package olive

import (
	"fmt"
	"log"
	"math/bits"
	"os"
	"strconv"
)

// Flag represents a flag that when encountered stores true
type Flag struct {
	name, shortName string
	desc            string
	action          func()
}

// Name gets the name of the flag
func (f *Flag) Name() string {
	return f.name
}

// ShortName gets the short name of the flag
func (f *Flag) ShortName() string {
	return f.shortName
}

// Description gets the description of the flag
func (f *Flag) Description() string {
	return f.desc
}

// SetAction sets an action function to be run if this flag is encountered
func (f *Flag) SetAction(fn func()) {
	f.action = fn
}

// -----------------------------------------------------------------------------

// Argument represents a value that can be passed to the application via a
// label (eg. `--loglevel=silent`).  There are many different kinds of arguments
// and so this is an interface to allow for sub-arguments
type Argument interface {
	// Name gives the full name of the argument
	Name() string

	// ShortName gives the short name of the argument
	ShortName() string

	// Description returns the description of the argument
	Description() string

	// Required indicates whether or not the argument is required
	Required() bool

	// GetDefaultValue gets the default value of the argument
	GetDefaultValue() (interface{}, bool)

	// checkValue is the function used by the parser to check argument values as
	// it collect them.  It returns an "any type" which contains the typed value
	// of the argument and an error indicating whether or not the argument value
	// was accepted
	checkValue(string) (interface{}, error)
}

// argumentBase is the base type for all special argument kinds
type argumentBase struct {
	name, shortName string
	desc            string
	required        bool
	defaultValue    interface{}
}

func (ab *argumentBase) Name() string {
	return ab.name
}

func (ab *argumentBase) ShortName() string {
	return ab.shortName
}

func (ab *argumentBase) Description() string {
	return ab.desc
}

func (ab *argumentBase) Required() bool {
	return ab.required
}

func (ab *argumentBase) GetDefaultValue() (interface{}, bool) {
	return ab.defaultValue, ab.defaultValue != nil
}

// IntArgument is an argument whose value must be an integer
type IntArgument struct {
	argumentBase

	validator func(int) error
}

// SetValidator sets a validation function for this argument
func (ia *IntArgument) SetValidator(v func(int) error) {
	ia.validator = v
}

// SetDefaultValue sets the default value of this argument
func (ia *IntArgument) SetDefaultValue(v int) {
	if ia.validator != nil {
		if err := ia.validator(v); err != nil {
			log.Fatalf("validator error: %s\n", err.Error())
		}
	}

	ia.defaultValue = v
}

func (ia *IntArgument) checkValue(val string) (interface{}, error) {
	// the int argument value is always the size of the default `int` type for
	// the platform (this should realistically never be an issue)
	raw, err := strconv.ParseInt(val, 0, bits.UintSize)
	if err != nil {
		return nil, err
	}

	v := int(raw)
	if ia.validator != nil {
		if err := ia.validator(v); err != nil {
			return nil, err
		}
	}

	return v, nil
}

// FloatArgument is an argument whose value must be a float
type FloatArgument struct {
	argumentBase

	validator func(float64) error
}

// SetValidator sets a validation function for this argument
func (fa *FloatArgument) SetValidator(v func(float64) error) {
	fa.validator = v
}

// SetDefaultValue sets the default value of this argument
func (fa *FloatArgument) SetDefaultValue(v float64) {
	if fa.validator != nil {
		if err := fa.validator(v); err != nil {
			log.Fatalf("validator error: %s\n", err.Error())
		}
	}

	fa.defaultValue = v
}

func (fa *FloatArgument) checkValue(val string) (interface{}, error) {
	v, err := strconv.ParseFloat(val, 64)

	if err != nil {
		return nil, err
	}

	if fa.validator != nil {
		if err := fa.validator(v); err != nil {
			return nil, err
		}
	}

	return v, nil
}

// StringArgument is an argument whose value must be a string
type StringArgument struct {
	argumentBase

	validator func(string) error
}

// SetValidator sets a validation function for this argument
func (sa *StringArgument) SetValidator(v func(string) error) {
	sa.validator = v
}

// SetDefaultValue sets the default value of this argument
func (sa *StringArgument) SetDefaultValue(v string) {
	if sa.validator != nil {
		if err := sa.validator(v); err != nil {
			log.Fatalf("validator error: %s\n", err.Error())
		}
	}

	sa.defaultValue = v
}

func (sa *StringArgument) checkValue(val string) (interface{}, error) {
	if sa.validator != nil {
		if err := sa.validator(val); err != nil {
			return nil, err
		}
	}

	return val, nil
}

// SelectorArgument is an argument whose value is constained to a finite set of
// string values
type SelectorArgument struct {
	argumentBase

	possibleValues map[string]struct{}
	validator      func(string) error
}

// SetValidator sets a validation function for this argument
func (sea *SelectorArgument) SetValidator(v func(string) error) {
	sea.validator = v
}

// SetDefaultValue sets the default value of this argument
func (sea *SelectorArgument) SetDefaultValue(v string) {
	_, err := sea.checkValue(v)
	if err != nil {
		log.Fatalf("default value error: %s\n", err.Error())
	}

	sea.defaultValue = v
}

func (sea *SelectorArgument) checkValue(val string) (interface{}, error) {
	if _, ok := sea.possibleValues[val]; !ok {
		return nil, fmt.Errorf("`%s` is not a valid value for argument [%s]", val, sea.name)
	}

	if sea.validator != nil {
		if err := sea.validator(val); err != nil {
			return nil, err
		}
	}

	return val, nil
}

// -----------------------------------------------------------------------------

// PrimaryArgument is an argument that is passed to command without an explicit
// label (eg. for `go build <filename>`, `<filename>` is the primary argument).
// Note that a command cannot both take a primary argument and subcommands.
type PrimaryArgument struct {
	name, desc string
}

// -----------------------------------------------------------------------------

func newCommand(name, desc string, helpEnabled bool) *Command {
	c := &Command{
		Name:               name,
		Description:        desc,
		subcommands:        make(map[string]*Command),
		flags:              make(map[string]*Flag),
		args:               make(map[string]Argument),
		flagsByShortName:   make(map[string]*Flag),
		argsByShortName:    make(map[string]Argument),
		RequiresSubcommand: true,
	}

	if helpEnabled {
		f := c.AddFlag("help", "h", "Get help")
		f.action = func() {
			c.Help()
			os.Exit(0)
		}
	}

	return c
}

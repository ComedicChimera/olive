package olive_test

import (
	"errors"
	"fmt"
	"log"
	"math"
	"olive"
	"os"
	"reflect"
	"testing"

	"bou.ke/monkey"
)

func TestCorrectFlags(t *testing.T) {
	cli := olive.NewCLI("olive", "", true)

	cli.AddFlag("flag1", "f1", "")

	f2 := cli.AddFlag("flag2", "f2", "")
	f2.SetAction(func() {
		t.Log("ran action")
	})

	cli.AddFlag("flag3", "f3", "")

	result, err := olive.ParseArgs(cli, []string{"olive", "-f1", "--flag2"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if !result.HasFlag("flag1") {
		t.Fatal("missing flag1")
	}

	if !result.HasFlag("flag2") {
		t.Fatal("missing flag2")
	}

	if result.HasFlag("flag3") {
		t.Fatal("flag3 set")
	}
}

func TestCorrectArgs(t *testing.T) {
	cli := olive.NewCLI("olive", "", true)

	cli.AddIntArg("int", "i", "", true)

	s := cli.AddStringArg("string", "s", "", false)
	s.SetDefaultValue("test")

	cli.AddFloatArg("float", "f", "", false)

	cli.AddSelectorArg("sel", "se", "", true, []string{"val1", "val2"})

	result, err := olive.ParseArgs(cli, []string{"olive", "--int=5", "-se=val1"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if result.Arguments["int"].(int) != 5 {
		t.Fatalf("expected value of `5` for argument `int`, not `%d`", result.Arguments["int"].(int))
	}

	if result.Arguments["string"].(string) != "test" {
		t.Fatalf("expected value of `test` for argument `string`, not `%s`", result.Arguments["string"].(string))
	}

	if result.Arguments["sel"].(string) != "val1" {
		t.Fatalf("expected value of `val1` for argument `sel`, not `%s`", result.Arguments["sel"].(string))
	}

	if _, ok := result.Arguments["float"]; ok {
		t.Fatalf("`float` should not have an argument value")
	}

	result, err = olive.ParseArgs(cli, []string{"olive", "-i=120", "--sel=val2", "-f=0.5", "-s=lul"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if result.Arguments["int"].(int) != 120 {
		t.Fatalf("expected value of `120` for argument `int`, not `%d`", result.Arguments["int"].(int))
	}

	if result.Arguments["string"].(string) != "lul" {
		t.Fatalf("expected value of `lul` for argument `string`, not `%s`", result.Arguments["string"].(string))
	}

	if result.Arguments["sel"].(string) != "val2" {
		t.Fatalf("expected value of `val2` for argument `sel`, not `%s`", result.Arguments["sel"].(string))
	}

	if result.Arguments["float"].(float64) != 0.5 {
		t.Fatalf("expected value of `0.5` for argument `float`, not `%f`", result.Arguments["float"].(float64))
	}
}

func TestCorrectFlagsandArgs(t *testing.T) {
	cli := olive.NewCLI("olive", "", true)

	cli.AddFlag("flag1", "f1", "")

	f2 := cli.AddFlag("flag2", "f2", "")
	f2.SetAction(func() {
		t.Log("ran action")
	})

	cli.AddIntArg("int", "i", "", true)

	s := cli.AddStringArg("string", "s", "", false)
	s.SetDefaultValue("test")

	cli.AddStringArg("flag1", "f1", "", true)

	result, err := olive.ParseArgs(cli, []string{"olive", "-f1", "--flag1=test", "-i=12"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if !result.HasFlag("flag1") {
		t.Fatal("missing flag: `flag1`")
	}

	if result.HasFlag("flag2") {
		t.Fatal("unexpected flag: `flag2`")
	}

	if result.Arguments["int"].(int) != 12 {
		t.Fatalf("expected value of `12` for argument `int`, not `%d`", result.Arguments["int"].(int))
	}

	if result.Arguments["string"].(string) != "test" {
		t.Fatalf("expected value of `test` for argument `string`, not `%s`", result.Arguments["string"].(string))
	}

	if result.Arguments["flag1"].(string) != "test" {
		t.Fatalf("expected value of `test` for argument `flag1`, not `%s`", result.Arguments["string"].(string))
	}

	result, err = olive.ParseArgs(cli, []string{"olive", "--flag2", "-f1=test", "--int=6", "-s=lul"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if !result.HasFlag("flag2") {
		t.Fatal("missing flag: `flag2`")
	}

	if result.HasFlag("flag1") {
		t.Fatal("unexpected flag: `flag1`")
	}

	if result.Arguments["int"].(int) != 6 {
		t.Fatalf("expected value of `6` for argument `int`, not `%d`", result.Arguments["int"].(int))
	}

	if result.Arguments["string"].(string) != "lul" {
		t.Fatalf("expected value of `lul` for argument `string`, not `%s`", result.Arguments["string"].(string))
	}

	if result.Arguments["flag1"].(string) != "test" {
		t.Fatalf("expected value of `test` for argument `flag1`, not `%s`", result.Arguments["string"].(string))
	}
}

func TestCorrectSubcommands(t *testing.T) {
	cli := olive.NewCLI("olive", "", true)

	cli.AddSubcommand("sub1", "", true)
	cli.AddSubcommand("sub2", "", true)

	c := cli.AddSubcommand("sub3", "", true)
	c.AddSubcommand("sub4", "", true)

	result, err := olive.ParseArgs(cli, []string{"olive", "sub1"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if name, _, ok := result.Subcommand(); ok {
		if name != "sub1" {
			t.Fatalf("unexpected subcommand on result: `%s`", name)
		}
	} else {
		t.Fatal("missing subcommand `sub1` on result")
	}

	result, err = olive.ParseArgs(cli, []string{"olive", "sub3", "sub4"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if name, subres, ok := result.Subcommand(); ok {
		if name != "sub3" {
			t.Fatalf("unexpected subcommand on result: `%s`", name)
		}

		if name, _, ok = subres.Subcommand(); ok {
			if name != "sub4" {
				t.Fatalf("unexpected subcommand on result: `%s`", name)
			}
		} else {
			t.Fatal("missing subcommand `sub4` on result")
		}
	} else {
		t.Fatal("missing subcommand `sub3` on result")
	}
}

func TestCorrectPrimaryArguments(t *testing.T) {
	cli := olive.NewCLI("olive", "", true)

	cli.AddSubcommand("subc1", "", true)

	c := cli.AddSubcommand("subc2", "", true)
	c.AddPrimaryArg("test", "")

	result, err := olive.ParseArgs(cli, []string{"olive", "subc1"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if name, _, ok := result.Subcommand(); ok {
		if name != "subc1" {
			t.Fatalf("unexpected subcommand on result: `%s`", name)
		}
	} else {
		t.Fatal("missing subcommand `subc1` on result")
	}

	result, err = olive.ParseArgs(cli, []string{"olive", "subc2", "val"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if name, res, ok := result.Subcommand(); ok {
		if name != "subc2" {
			t.Fatalf("unexpected subcommand on result: `%s`", name)
		}

		if primVal, ok := res.PrimaryArg(); ok {
			if primVal != "val" {
				t.Fatalf("unexpected primary argument: `%s`", primVal)
			}
		} else {
			t.Fatal("missing primary argument for command `subc2`")
		}
	} else {
		t.Fatal("missing subcommand `subc2` on result")
	}

	t.Log(c.HelpMessage())
}

func TestOptionalSubcommand(t *testing.T) {
	cli := olive.NewCLI("olive", "", true)

	cli.RequiresSubcommand = false
	cli.AddSubcommand("subc", "", true)

	_, err := olive.ParseArgs(cli, []string{"olive"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	result, err := olive.ParseArgs(cli, []string{"olive", "subc"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if name, _, ok := result.Subcommand(); ok {
		if name != "subc" {
			t.Fatalf("unexpected subcommand: `%s`", name)
		}
	} else {
		t.Fatal("missing subcommand `subc`")
	}
}

func TestCorrectMixedCLI(t *testing.T) {
	cli := olive.NewCLI("olive", "", true)

	cli.AddFlag("verbose", "v", "")

	cli.AddSubcommand("version", "", true)

	c := cli.AddSubcommand("build", "", true)
	c.AddPrimaryArg("package-name", "")
	c.AddStringArg("profile", "p", "", false)
	s := c.AddStringArg("output", "o", "", true)
	s.SetDefaultValue("cool_path")

	c2 := cli.AddSubcommand("mod", "", true)
	c3 := c2.AddSubcommand("init", "", true)
	c3.AddPrimaryArg("module-name", "")
	c2.AddSubcommand("update", "", true)

	result, err := olive.ParseArgs(cli, []string{"olive", "build", "-o=other_path", "package"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if name, res, ok := result.Subcommand(); ok {
		if name != "build" {
			t.Fatalf("unexpected subcommand: %s", name)
		}

		if res.Arguments["output"].(string) != "other_path" {
			t.Fatalf("expected argument value of `other_path` not `%s`", res.Arguments["output"].(string))
		}

		if primVal, ok := res.PrimaryArg(); ok {
			if primVal != "package" {
				t.Fatalf("expected primary argument value of `package` not `%s`", primVal)
			}
		} else {
			t.Fatal("missing primary argument on subcommand `build`")
		}
	} else {
		t.Fatal("missing subcommand `build`")
	}

	result, err = olive.ParseArgs(cli, []string{"olive", "mod", "init", "-v", "pog"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if !result.HasFlag("verbose") {
		t.Fatal("missing flag `verbose`")
	}

	if name, res, ok := result.Subcommand(); ok {
		if name != "mod" {
			t.Fatalf("unexpected subcommand: `%s`", name)
		}

		if name, subres, ok := res.Subcommand(); ok {
			if name != "init" {
				t.Fatalf("unexpected subcommand: `%s`", name)
			}

			if val, ok := subres.PrimaryArg(); ok {
				if val != "pog" {
					t.Fatalf("expected primary argument of `pog` not `%s`", val)
				}
			} else {
				t.Fatal("missing primary argument to `init`")
			}
		} else {
			t.Fatal("missing subcommand `init`")
		}
	} else {
		t.Fatal("missing subcommand `mod`")
	}

	t.Log(cli.HelpMessage())
}

func TestBadInput(t *testing.T) {
	cli := olive.NewCLI("olive", "", true)

	cli.AddFlag("verbose", "v", "")

	cli.AddSubcommand("version", "", true)

	c := cli.AddSubcommand("build", "", true)
	c.AddPrimaryArg("package-name", "")
	c.AddStringArg("profile", "p", "", false)
	s := c.AddStringArg("output", "o", "", true)
	s.SetDefaultValue("cool_path")

	c2 := cli.AddSubcommand("mod", "", true)
	c3 := c2.AddSubcommand("init", "", true)
	c3.AddPrimaryArg("module-name", "")
	c3.AddFlag("flag", "f", "")
	c2.AddSubcommand("update", "", true)
	c2.AddIntArg("int", "i", "", true)

	_, err := olive.ParseArgs(cli, []string{"olive"})
	if err == nil {
		t.Fatalf("missing subc error")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "-f"})
	if err == nil {
		t.Fatalf("missing flag error")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "subc"})
	if err == nil {
		t.Fatalf("missing unknown subc error")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "mod", "-i=10.5"})
	if err == nil {
		t.Fatalf("missing int flag error")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "mod", "init", "-int=10"})
	if err == nil {
		t.Fatalf("missing no primary arg value error")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "-f", "mod"})
	if err == nil {
		t.Fatal("missing unexpected subcommand error")
	}
}

func TestBadInput2(t *testing.T) {
	cli := olive.NewCLI("olive", "", true)

	cli.AddPrimaryArg("primary", "")
	cli.AddFlag("flag1", "f1", "")
	cli.AddSelectorArg("sel", "s", "", true, []string{"val1", "val2", "val3"})

	_, err := olive.ParseArgs(cli, []string{"olive", "prim1", "prim2"})
	if err == nil {
		t.Fatal("missing multiple primary arguments error")
	}

	result, err := olive.ParseArgs(cli, []string{"olive", "-f1", "prim"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if !result.HasFlag("flag1") {
		t.Fatal("missing flag `flag1`")
	}

	if val, ok := result.PrimaryArg(); ok {
		if val != "prim" {
			t.Fatalf("unexpected primary argument value: `%s`", val)
		}
	} else {
		t.Fatal("missing primary argument")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "-f1", "--flag1"})
	if err == nil {
		t.Fatal("missing flag set multiple times error")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "--sel=val1 -s=val2"})
	if err == nil {
		t.Fatal("missing arg set multiple times error")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "-s=val4"})
	if err == nil {
		t.Fatal("missing invalid selection error")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "--v"})
	if err == nil {
		t.Fatal("missing unknown flag error")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "--f=10.2"})
	if err == nil {
		t.Fatal("missing unknown argument error")
	}
}

func TestHelp(t *testing.T) {
	monkey.Patch(os.Exit, func(int) {
		t.Log("help exited application")
	})

	defer monkey.Unpatch(os.Exit)

	monkey.Patch(fmt.Println, func(a ...interface{}) (int, error) {
		t.Log("displaying help")
		return 0, nil
	})

	defer monkey.Unpatch(fmt.Println)

	cli := olive.NewCLI("olive", "", true)

	result, err := olive.ParseArgs(cli, []string{"olive", "-h"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if !result.HasFlag("help") {
		t.Fatal("missing help flag")
	}

	cli.DisableHelp()

	_, err = olive.ParseArgs(cli, []string{"olive", "--help"})
	if err == nil {
		t.Fatal("missing unknown flag error")
	}

	cli.EnableHelp()

	result, err = olive.ParseArgs(cli, []string{"olive", "-h"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if !result.HasFlag("help") {
		t.Fatal("missing help flag")
	}

	cli2 := olive.NewCLI("olive2", "", false)

	_, err = olive.ParseArgs(cli2, []string{"olive", "--help"})
	if err == nil {
		t.Fatal("missing unknown flag error")
	}

	cli2.EnableHelp()

	result, err = olive.ParseArgs(cli2, []string{"olive", "-h"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if !result.HasFlag("help") {
		t.Fatal("missing help flag")
	}
}

func TestBadConfig(t *testing.T) {
	logFatalCount := 0

	monkey.Patch(log.Fatalf, func(format string, v ...interface{}) {
		t.Log(format)
		logFatalCount++
	})

	defer monkey.Unpatch(log.Fatalf)

	// we don't technically care if the CLI object gets screwed up -- just that
	// we proked the error
	cli := olive.NewCLI("olive", "", true)

	cli.AddFlag("help", "he", "") // fatal 1
	cli.AddFlag("he", "h", "")    // fatal 2

	cli.AddIntArg("int", "i", "", true)
	cli.AddFloatArg("int", "in", "", true)    // fatal 3
	cli.AddStringArg("string", "i", "", true) // fatal 4

	cli.AddPrimaryArg("p", "")

	cli.AddSubcommand("cheeky", "", true) // fatal 5

	cli = olive.NewCLI("olive2", "", true)
	cli.AddSubcommand("bug", "", true)

	cli.AddSubcommand("bug", "", true) // fatal 6

	cli.AddPrimaryArg("b", "") // fatal 7

	if logFatalCount != 7 {
		t.Fatalf("expected 7 fatal errors: received %d", logFatalCount)
	}
}

func TestValidatorsandDefaults(t *testing.T) {
	cli := olive.NewCLI("olive", "", false)

	ia := cli.AddIntArg("int", "i", "", false)
	ia.SetValidator(func(x int) error {
		if x%2 == 1 {
			return errors.New("must be even")
		}

		return nil
	})
	ia.SetDefaultValue(0)

	fa := cli.AddFloatArg("float", "f", "", false)
	fa.SetValidator(func(x float64) error {
		if math.Mod(x, 1) >= 0.5 {
			return errors.New("must round down")
		}

		return nil
	})
	fa.SetDefaultValue(0.2)

	sa := cli.AddStringArg("str", "s", "", false)
	sa.SetValidator(func(x string) error {
		if len(x) > 5 {
			return errors.New("must be shorter than 6 chars")
		}

		return nil
	})
	sa.SetDefaultValue("val")

	sea := cli.AddSelectorArg("sel", "se", "", false, []string{"val1", "val2", "badVal"})
	sea.SetValidator(func(x string) error {
		if x == "badVal" {
			return errors.New("bad val")
		}

		return nil
	})
	sea.SetDefaultValue("val1")

	// all defaults
	result, err := olive.ParseArgs(cli, []string{"olive"})
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}

	if !reflect.DeepEqual(result.Arguments, map[string]interface{}{
		"int":   0,
		"float": 0.2,
		"str":   "val",
		"sel":   "val1",
	}) {
		t.Fatalf("bad default argument fill ins")
	}

	result, err = olive.ParseArgs(cli, []string{"olive", "-i=2", "-f=0.3", "-s=v", "-se=val2"})
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}

	if !reflect.DeepEqual(result.Arguments, map[string]interface{}{
		"int":   2,
		"float": 0.3,
		"str":   "v",
		"sel":   "val2",
	}) {
		t.Fatalf("bad argument fill ins")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "-i=5"})
	if err == nil {
		t.Fatalf("Int validator didn't catch")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "-f=0.8"})
	if err == nil {
		t.Fatalf("Float validator didn't catch")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "-s=abcdef"})
	if err == nil {
		t.Fatalf("String validator didn't catch")
	}

	_, err = olive.ParseArgs(cli, []string{"olive", "-se=badVal"})
	if err == nil {
		t.Fatalf("Selector validator didn't catch")
	}
}

func TestDisplayInterf(t *testing.T) {
	cli := olive.NewCLI("olive", "", false)

	f := cli.AddFlag("flag", "f", "Test description")
	if f.Name() != "flag" {
		t.Fatalf("Flag should have name `flag` not `%s`", f.Name())
	}

	if f.ShortName() != "f" {
		t.Fatalf("Flag should have short name `f` not `%s`", f.ShortName())
	}

	if f.Description() != "Test description" {
		t.Fatalf("Flag should have description `Test description` not `%s`", f.Description())
	}

	a := cli.AddIntArg("int", "i", "Int description", true)
	if a.Name() != "int" {
		t.Fatalf("Argument should have name `int` not `%s`", f.Name())
	}

	if a.ShortName() != "i" {
		t.Fatalf("Argument should have short name `i` not `%s`", f.ShortName())
	}

	if a.Description() != "Int description" {
		t.Fatalf("Argument should have description `Int description` not `%s`", f.Description())
	}

	if !a.Required() {
		t.Fatal("Argument should be marked as required")
	}
}

func TestBadDefaultValues(t *testing.T) {
	logFatalCount := 0

	monkey.Patch(log.Fatalf, func(format string, v ...interface{}) {
		t.Log(format)
		logFatalCount++
	})

	defer monkey.Unpatch(log.Fatalf)

	cli := olive.NewCLI("olive", "", false)

	ia := cli.AddIntArg("int", "i", "", false)
	ia.SetValidator(func(x int) error {
		if x%2 == 1 {
			return errors.New("must be even")
		}

		return nil
	})
	ia.SetDefaultValue(1) // fatal 1

	fa := cli.AddFloatArg("float", "f", "", false)
	fa.SetValidator(func(x float64) error {
		if math.Mod(x, 1) >= 0.5 {
			return errors.New("must round down")
		}

		return nil
	})
	fa.SetDefaultValue(0.6) // fatal 2

	sa := cli.AddStringArg("str", "s", "", false)
	sa.SetValidator(func(x string) error {
		if len(x) > 5 {
			return errors.New("must be shorter than 6 chars")
		}

		return nil
	})
	sa.SetDefaultValue("abcdef") // fatal 3

	sea := cli.AddSelectorArg("sel", "se", "", false, []string{"val1", "val2", "badVal"})
	sea.SetValidator(func(x string) error {
		if x == "badVal" {
			return errors.New("bad val")
		}

		return nil
	})
	sea.SetDefaultValue("badVal") // fatal 4

	if logFatalCount != 4 {
		t.Fatalf("expected `4` fatal errors; received `%d`", logFatalCount)
	}
}

# Olive

[![Go Report Card](https://goreportcard.com/badge/github.com/ComedicChimera/olive)](https://goreportcard.com/report/github.com/ComedicChimera/olive)
[![Build Status](https://travis-ci.org/ComedicChimera/olive.svg?branch=master)](https://travis-ci.org/ComedicChimera/olive)

Olive is a delightful, little argument parsing library for Golang designed to
replace the builtin *flag* package.  Olive is lightweight, easy to use, and
intuitive while still being powerful enough and customizable enough to suit
your needs.

## Installation

To get started using Olive, enter the following command into your terminal of
choice.  

    go get -u github.com/ComedicChimera/olive

Note that Olive uses Go modules -- you will need to make sure you are on a
version of Go that supports modules.

## Quickstart

Once you have installed Olive, using it is easy.  Here is a quick sample of
parsing command line arguments for a simple application.

```go
import (
    "fmt"
    "os"
    "errors"

    "github.com/ComedicChimera/olive"
)

func main() {
    cli := olive.NewCLI("sample", "A sample command line utility for Olive")

    cli.AddFlag("show-expr", "se", help: "Show the expression computed")

    cli.AddIntArg("value1", "v1", help: "The first value", required: true)
    
    argv2 := cli.AddIntArg("value2", "v2", help: "The second value", required: true)
    argv2.SetValidator(func(v int) error {
        if v != 0 {
            return errors.New("cannot divide by zero")
        }

        return nil
    })

    result, err := olive.ParseArgs(cli, os.Args)
    if err != nil {
        fmt.Println(err)
        return
    }

    v1 := result.Arguments["value1"].(int)
    v2 := result.Arguments["value2"].(int)

    if result.GetFlag("show-expr") {
        fmt.Printf("%d / %d ", v1, v2)
    }

    fmt.Println("=", v1 / v2)
}
```

A sample call of this program would look like:

    > sample -v1=10 -v2=2 --show-expr
    10 / 2 = 5

## Documentation

For more specific information on how to use Olive, check out the [wiki](https://github.com/ComedicChimera/olive/wiki).

## Contributing

If you would like to contribute to this repository, there are a number of things
that could be worked on/need doing.  Just fork this repo and make a pull request
once you have finished modifying it.

- Adding more tests
- Adding additional argument types
- Adding constraints and/or validators to primary arguments
- Improving/adding to documentation

If you encounter any bugs or think any additional features that this repo should
have, but don't have the time to implement them yourself, feel free to open an
issue.

Note that Olive is not really my primary project atm (I basically made it to
fill a need I had in another, more serious project) so there may be a bit of a
delay in feature releases.



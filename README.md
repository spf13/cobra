# Cobra

A Commander for modern go CLI interactions

## Overview
Cobra provides a simple interface to create powerful modern CLI
interfaces similar to git & go tools.

Cobra was inspired by go, go-Commander, gh and subcommand


## Concepts

There are 3 different core objects to become familiar with to use Cobra.
To help illustrate these 3 items better use the following as an example:

    hugo server --port=1313

### Commander

The Commander is the head of your application. It holds the configuration
for your application. It also is responsible for all global flags.

In the example above 'hugo' is the commander.


### Command

Command is the central point of the application. Each interaction that
the application supports will be contained in a Command. A command can
have children commands and optionally run an action.

In the example above 'server' is the command


### Flags

A flag is a way to modify the behavior of an command. Cobra supports
fully posix compliant flags as well as remaining consistent with
the go flag package. A Cobra command has can define flags that 
persist through to children commands and flags that are only available
to that command.

In the example above 'port' is the flag.

## Usage

### Implementing Cobra

Using Cobra is easy. First use go get to install the latest version
of the library.

    $ go get github.com/spf13/cobra

Next include cobra in your application.

    import "github.com/spf13/cobra"


While it may be counter intuitive, You define your commands first,
assign flags to them, add them to the commander and lastly
execute the commander.

### Examples


## Simple Example

    Import(
        "github.com/spf13/cobra"
        "fmt"
        "strings"
    )

    func main() {

        var echoTimes int

        var cmdPrint = &cobra.Command{
            Use:   "print [string to print]",
            Short: "Print anything to the screen",
            Long:  `an utterly useless command for testing.`,
            Run: func(cmd *cobra.Command, args []string) {
                fmt.Println("Print: " + strings.Join(args, " "))
            },
        }

        var cmdEcho = &cobra.Command{
            Use:   "echo [string to echo]",
            Short: "Echo anything to the screen",
            Long:  `an utterly useless command for testing.`,
            Run: func(cmd *cobra.Command, args []string) {
                for i:=0; i < echoTimes; i++ {
                    fmt.Println("Echo: " + strings.Join(args, " "))
                }
            },
        }


        cmdEcho.Flags().IntVarP(&echoTimes, "times", "t", 1, "times to echo the input")

        var commander = cobra.Commander()
        commander.SetName("Cobra")
        commander.AddCommand(cmdPrint, cmdEcho)
        commander.Execute()
    }




## Release Notes

* **0.1.0** Sept 3, 2013
  * Implement first draft

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

## Contributors

Names in no particular order:

* [spf13](https://github.com/spf13)

## License

nitro is released under the Apache 2.0 license. See [LICENSE.txt](https://github.com/spf13/nitro/blob/master/LICENSE.txt)

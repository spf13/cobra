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

Now you are ready to implement Cobra. 

Cobra works by creating a set of commands and then organizing them into a tree.
The tree defines the structure of the application.

Once each command is defined with it's corresponding flags, then the
tree is assigned to the commander which is finally executed.

In the example below we have defined three commands. Two are at the top
level and one (cmdTimes) is a child of one of the top commands.

We have only defined one flag for a single command.

More documentation about flags is available at https://github.com/spf13/pflag

## Example

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
            Long:  `print is for printing anything back to the screen.
            For many years people have printed back to the screen.
            `,
            Run: func(cmd *cobra.Command, args []string) {
                fmt.Println("Print: " + strings.Join(args, " "))
            },
        }

        var cmdEcho = &cobra.Command{
            Use:   "echo [string to echo]",
            Short: "Echo anything to the screen",
            Long:  `echo is for echoing anything back.
            Echo works a lot like print, except it has a child command.
            `,
            Run: func(cmd *cobra.Command, args []string) {
                fmt.Println("Print: " + strings.Join(args, " "))
            },
        }

        var cmdTimes = &cobra.Command{
            Use:   "times [# times] [string to echo]",
            Short: "Echo anything to the screen more times",
            Long:  `echo things multiple times back to the user by providing
            a count and a string.`,
            Run: func(cmd *cobra.Command, args []string) {
                for i:=0; i < echoTimes; i++ {
                    fmt.Println("Echo: " + strings.Join(args, " "))
                }
            },
        }


        cmdTimes().IntVarP(&echoTimes, "times", "t", 1, "times to echo the input")

        var commander = cobra.Commander()
        commander.SetName("CobraExample")
        commander.AddCommand(cmdPrint, cmdEcho)
        cmdEcho.AddCommand(cmdTimes)
        commander.Execute()
    }

## Release Notes
* **0.7.0** Sept 24, 2013
  * Needs more eyes
  * Test suite
  * Support for automatic error messages
  * Support for help command
  * Support for printing to any io.Writer instead of os.Stderr
  * Support for persistent flags which cascade down tree
  * Ready for integration into Hugo
* **0.1.0** Sept 3, 2013
  * Implement first draft

## ToDo
* More testing of non-runnable
* More testing
* Launch proper documentation site

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

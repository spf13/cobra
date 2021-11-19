# Cobra Generator

Cobra provides its own program that will create your application and add any
commands you want. It's the easiest way to incorporate Cobra into your application.

Install the cobra generator with the command `go install github.com/spf13/cobra/cobra`. 
Go will automatically install it in your `$GOPATH/bin` directory which should be in your $PATH. 

Once installed you should have the `cobra` command available. Confirm by typing `cobra` at a 
command line. 

There are only two operations currently supported by the Cobra generator: 

### cobra init

The `cobra init [app]` command will create your initial application code
for you. It is a very powerful application that will populate your program with
the right structure so you can immediately enjoy all the benefits of Cobra. 
It can also apply the license you specify to your application.

With the introduction of Go modules, the Cobra generator has been simplified to
take advantage of modules. The Cobra generator works from within a Go module. 

#### Initalizing a module

__If you already have a module, skip this step.__

If you want to initialize a new Go module: 

 1. Create a new directory 
 2. `cd` into that directory
 3. run `go mod init <MODNAME>`

e.g. 
```
cd $HOME/code 
mkdir myapp
cd myapp
go mod init github.com/spf13/myapp
```

#### Initalizing an Cobra CLI application

From within a Go module run `cobra init`. This will create a new barebones project
for you to edit. 

You should be able to run your new application immediately. Try it with 
`go run main.go`. 

You will want to open up and edit 'cmd/root.go' and provide your own description and logic. 

e.g.
```
cd $HOME/code/myapp
cobra init
go run main.go
```

Cobra init can also be run from a subdirectory such as how the [cobra generator itself is organized](https://github.com/spf13/cobra).
This is useful if you want to keep your application code separate from your library code.

#### Optional flags:
You can provide it your author name with the `--author` flag. 
e.g. `cobra init --author "Steve Francia spf@spf13.com"`

You can provide a license to use with `--license` 
e.g. `cobra init --license apache`

Use the `--viper` flag to automatically setup [viper](https://github.com/spf13/viper)

Viper is a companion to Cobra intended to provide easy handling of environment variables and config files and seamlessly connecting them to the application flags.

### Add commands to a project

Once a cobra application is initialized you can continue to use the Cobra generator to 
add additional commands to your application. The command to do this is `cobra add`. 

Let's say you created an app and you wanted the following commands for it:

* app serve
* app config
* app config create

In your project directory (where your main.go file is) you would run the following:

```
cobra add serve
cobra add config
cobra add create -p 'configCmd'
```

`cobra add` supports all the same optional flags as `cobra init` does (described above).

You'll notice that this final command has a `-p` flag. This is used to assign a
parent command to the newly added command. In this case, we want to assign the
"create" command to the "config" command. All commands have a default parent of rootCmd if not specified.  

By default `cobra` will append `Cmd` to the name provided and uses this name for the internal variable name. When specifying a parent, be sure to match the variable name used in the code. 

*Note: Use camelCase (not snake_case/kebab-case) for command names.
Otherwise, you will encounter errors.
For example, `cobra add add-user` is incorrect, but `cobra add addUser` is valid.*

Once you have run these three commands you would have an app structure similar to
the following:

```
  ▾ app/
    ▾ cmd/
        config.go
        create.go
        serve.go
        root.go
      main.go
```

At this point you can run `go run main.go` and it would run your app. `go run
main.go serve`, `go run main.go config`, `go run main.go config create` along
with `go run main.go help serve`, etc. would all work.

You now have a basic Cobra-based application up and running. Next step is to edit the files in cmd and customize them for your application.

For complete details on using the Cobra library, please read the [The Cobra User Guide](https://github.com/spf13/cobra/blob/master/user_guide.md#using-the-cobra-library).

Have fun!

### Configuring the cobra generator

The Cobra generator will be easier to use if you provide a simple configuration
file which will help you eliminate providing a bunch of repeated information in
flags over and over.

An example ~/.cobra.yaml file:

```yaml
author: Steve Francia <spf@spf13.com>
license: MIT
viper: true
```

You can also use built-in licenses. For example, **GPLv2**, **GPLv3**, **LGPL**,
**AGPL**, **MIT**, **2-Clause BSD** or **3-Clause BSD**.

You can specify no license by setting `license` to `none` or you can specify
a custom license:

```yaml
author: Steve Francia <spf@spf13.com>
year: 2020
license:
  header: This file is part of CLI application foo.
  text: |
    {{ .copyright }}

    This is my license. There are many like it, but this one is mine.
    My license is my best friend. It is my life. I must master it as I must
    master my life.
```

In the above custom license configuration the `copyright` line in the License
text is generated from the `author` and `year` properties. The content of the
`LICENSE` file is

```
Copyright © 2020 Steve Francia <spf@spf13.com>

This is my license. There are many like it, but this one is mine.
My license is my best friend. It is my life. I must master it as I must
master my life.
```

The `header` property is used as the license header files. No interpolation is
done. This is the example of the go file header.
```
/*
Copyright © 2020 Steve Francia <spf@spf13.com>
This file is part of CLI application foo.
*/
```

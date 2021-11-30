package main

import (
    "os"
    "github.com/viscory/reciprocus/cli"
)

func main() {
    defer os.Exit(0)

    cli := cli.CommandLine{}
    cli.Run()
}

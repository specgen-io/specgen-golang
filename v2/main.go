package main

import (
	"fmt"
	"github.com/specgen-io/specgen-golang/v2/generators"
	"github.com/specgen-io/specgen-golang/v2/goven/generator"
	"github.com/specgen-io/specgen-golang/v2/goven/generator/console"
	"github.com/specgen-io/specgen-golang/v2/version"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "specgen",
		Version: version.Current,
		Short:   "Code generation based on specification",
	}
	generator.AddCobraCommands(rootCmd, generators.All)
	cobra.OnInitialize()
	console.PrintLn("Running specgen")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

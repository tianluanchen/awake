package cmd

import (
	"awake/pkg"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "awake",
	Short: "A toolkit",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		level, _ := cmd.Flags().GetString("level")
		level = strings.ToUpper(level)
		switch level {
		case "DEBUG":
			pkg.SetLogLevel(pkg.LDEBUG)
		case "INFO":
			pkg.SetLogLevel(pkg.LINFO)
		case "WARN":
			pkg.SetLogLevel(pkg.LWARN)
		case "ERROR":
			pkg.SetLogLevel(pkg.LERROR)
		case "FATAL":
			pkg.SetLogLevel(pkg.LFATAL)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().String("level", "INFO", "log level, DEBUG INFO WARN ERROR FATAL")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

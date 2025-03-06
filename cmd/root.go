package cmd

import (
	"awake/pkg"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var Version = "0.0.0"

var rootCmd = &cobra.Command{
	Use:     "awake",
	Version: Version,
	Short:   "A toolkit",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		level, _ := cmd.Flags().GetString("level")
		level = strings.ToUpper(level)
		pkg.SetLogOutput(os.Stdout)
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

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

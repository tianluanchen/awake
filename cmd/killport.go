package cmd

import (
	"awake/pkg"
	"errors"
	"strconv"

	"github.com/spf13/cobra"
)

var killPortCmd = &cobra.Command{
	Use:   "killport",
	Short: "Kill processes occupying local ports",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		kind, _ := cmd.Flags().GetString("kind")
		switch kind {
		case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6", "inet", "inet4", "inet6":
		default:
			logger.Fatalln("kind must be one of tcp, tcp4, tcp6, udp, udp4, udp6, inet, inet4, inet6")
		}
		ports := make([]int, 0)
		portMap := make(map[int]struct{})
		for i := range len(args) {
			p, err := strconv.Atoi(args[i])
			if err == nil && (p < 0 || p > 65535) {
				err = errors.New("port must be between 0 and 65535")
			}
			if err != nil {
				logger.Fatalf("%s is not a valid port: %v", args[i], err)
			}
			if _, ok := portMap[p]; !ok {
				portMap[p] = struct{}{}
				ports = append(ports, p)
			}
		}
		for _, p := range ports {
			err := pkg.KillPortProcess(p, kind)
			if err != nil {
				logger.Fatalf("kill kind %s port %d error: %v", kind, p, err)
			}
		}
	},
}

func init() {
	killPortCmd.Flags().StringP("kind", "k", "tcp", "kind must be one of tcp, tcp4, tcp6, udp, udp4, udp6, inet, inet4, inet6")
	rootCmd.AddCommand(killPortCmd)
}

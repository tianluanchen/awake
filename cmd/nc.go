package cmd

import (
	"io"
	"net"
	"os"
	"sync"

	"github.com/spf13/cobra"
)

var ncCmd = &cobra.Command{
	Use:     "nc",
	Short:   "Netcat, only for tcp",
	Example: "  awake nc 1.1.1.1 80 -p socks5://127.0.0.1:1080",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		targetAddr := net.JoinHostPort(args[0], args[1])
		proxy, _ := cmd.Flags().GetString("proxy")
		var (
			conn net.Conn
			err  error
		)
		if proxy == "" {
			conn, err = net.Dial("tcp", targetAddr)
		} else {
			conn, err = dialTCPWithProxy(proxy, targetAddr)
		}
		if err != nil {
			logger.Fatalln(err)
		}
		var once sync.Once
		ch := make(chan error, 1)
		go func() {
			_, err = io.Copy(conn, os.Stdin)
			once.Do(func() {
				ch <- err
			})
		}()
		go func() {
			_, err = io.Copy(os.Stdout, conn)
			once.Do(func() {
				ch <- err
			})
		}()
		err = <-ch
		conn.Close()
		if err != nil {
			logger.Fatalln(err)
		}
	},
}

func init() {
	ncCmd.Flags().StringP("proxy", "p", "", "proxy url")
	rootCmd.AddCommand(ncCmd)
}

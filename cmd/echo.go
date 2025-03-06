package cmd

import (
	"io"
	"net"
	"sync"

	"github.com/spf13/cobra"
)

var echoCmd = &cobra.Command{
	Use:   "echo",
	Short: "Start tcp/udp echo server",
	Run: func(cmd *cobra.Command, args []string) {
		addr, _ := cmd.Flags().GetString("addr")
		udp, _ := cmd.Flags().GetBool("udp")
		tcp, _ := cmd.Flags().GetBool("tcp")
		var wg sync.WaitGroup
		if udp {
			wg.Add(1)
			go func() {
				defer wg.Done()
				laddr, err := net.ResolveUDPAddr("udp", addr)
				if err != nil {
					logger.Fatalln(err)
				}
				conn, err := net.ListenUDP("udp", laddr)
				if err != nil {
					logger.Fatalln(err)
				}
				logger.Infoln("udp echo server listen on", laddr.String())
				buf := make([]byte, 65535)
				for {
					n, raddr, err := conn.ReadFromUDP(buf)
					if err != nil {
						logger.Fatalln(err)
					}
					var temp string
					if n > 64 {
						temp = string(buf[:64]) + "..."
					} else {
						temp = string(buf[:n])
					}
					logger.Infof("[udp] received %d bytes from %s: %s", n, raddr.String(), temp)
					_, err = conn.WriteToUDP(buf[:n], raddr)
					if err != nil {
						logger.Fatalln(err)
					} else {
						logger.Infof("[udp] replied %d bytes to %s", n, raddr.String())
					}
				}
			}()
		}
		if tcp {
			wg.Add(1)
			go func() {
				defer wg.Done()
				laddr, err := net.ResolveTCPAddr("tcp", addr)
				if err != nil {
					logger.Fatalln(err)
				}
				conn, err := net.ListenTCP("tcp", laddr)
				if err != nil {
					logger.Fatalln(err)
				}
				logger.Infoln("tcp echo server listen on", laddr.String())
				for {
					conn, err := conn.AcceptTCP()
					if err != nil {
						logger.Fatalln(err)
					}
					logger.Infof("[tcp] accepted connection from %s", conn.RemoteAddr())
					go func() {
						for {
							buf := make([]byte, 1024*32)
							n, err := conn.Read(buf)
							if err != nil {
								if err == io.EOF {
									logger.Infof("[tcp] connection closed by %s", conn.RemoteAddr())
								} else {
									logger.Warnln(err)
								}
								return
							}
							var temp string
							if n > 64 {
								temp = string(buf[:64]) + "..."
							} else {
								temp = string(buf[:n])
							}
							logger.Infof("[tcp] received %d bytes from %s: %s", n, conn.RemoteAddr(), temp)
							_, err = conn.Write(buf[:n])
							if err != nil {
								logger.Warnln(err)
								return
							}
							logger.Infof("[tcp] replied %d bytes to %s", n, conn.RemoteAddr())
						}
					}()
				}
			}()
		}
		wg.Wait()
	},
}

func init() {
	echoCmd.Flags().StringP("addr", "a", "127.0.0.1:8080", "listen address")
	echoCmd.Flags().Bool("udp", false, "start udp echo server")
	echoCmd.Flags().Bool("tcp", true, "start tcp echo server")
	rootCmd.AddCommand(echoCmd)
}

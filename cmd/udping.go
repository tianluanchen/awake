package cmd

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var udpingCmd = &cobra.Command{
	Use:   "udping",
	Short: "Udping",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, _ := cmd.Flags().GetInt("count")
		s, _ := cmd.Flags().GetString("string")
		interval, _ := cmd.Flags().GetDuration("interval")
		isHex, _ := cmd.Flags().GetBool("hex")
		var data []byte
		if isHex {
			b, err := hex.DecodeString(s)
			if err != nil {
				return err
			}
			data = b
		} else {
			data = []byte(s)
		}
		if interval < time.Millisecond*250 {
			return errors.New("interval too small, minimum 250ms")
		}
		if u, err := url.Parse(args[0]); err == nil {
			host, port := u.Hostname(), u.Port()
			if port == "" {
				if u.Scheme == "https" {
					port = "443"
				} else if u.Scheme == "http" {
					port = "80"
				}
			}
			if host != "" && port != "" {
				args[0] = net.JoinHostPort(host, port)
			}
		}
		udpAddr, err := net.ResolveUDPAddr("udp", args[0])
		if err != nil {
			return err
		}
		addr := udpAddr.String()
		fmt.Printf("Udpinging %s (%s) :\n", args[0], addr)
		buf := make([]byte, 1024)
		infinite := n <= 0
		i := 0
		var total, success int64
		var min, max, sum time.Duration
		printStats := func() {
			var avg time.Duration
			if success != 0 {
				avg = sum / time.Duration(success)
			}
			fmt.Printf("\nTotal = %d, Success = %d, Fail = %d, Pass Percentage = %.1f%%\nMin = %v, Max = %v, Avg = %v\n",
				total, success, total-success, float64(success)/float64(total)*100, min, max, avg)
		}
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT)
		go func() {
			<-sigChan
			printStats()
			os.Exit(0)
		}()
		defer printStats()
		for {
			start := time.Now()
			conn, err := net.Dial("udp", addr)
			if err == nil {
				_, err = conn.Write(data)
				if err == nil {
					err = conn.SetReadDeadline(time.Now().Add(time.Second * 6))
					if err == nil {
						readN := 0
						readN, err = conn.Read(buf)
						if err == nil {
							t := time.Since(start)
							conn.Close()
							fmt.Printf("Reply from %s : time=%s  content(%d)=%s\n", addr, t, readN, string(buf[:readN]))
							success += 1
							sum += t
							if min == 0 || t < min {
								min = t
							}
							if t > max {
								max = t
							}
						}
					}
				}
			}
			total += 1
			if err != nil {
				fmt.Printf("Unexpected error: %s\n", err)
			}
			i++
			if infinite || i < n {
				time.Sleep(interval)
			} else {
				return nil
			}
		}
	},
}

func init() {
	udpingCmd.Flags().StringP("string", "s", "ping", "the string to be sent")
	udpingCmd.Flags().Bool("hex", false, "input as hexadecimal string and decode it")
	udpingCmd.Flags().DurationP("interval", "i", time.Second, "time between sending each packet, minimum 400ms")
	udpingCmd.Flags().IntP("count", "c", 3, "ping times, nonpositive number means infinity")
	rootCmd.AddCommand(udpingCmd)
}

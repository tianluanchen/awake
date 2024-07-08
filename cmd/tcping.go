package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var tcpingCmd = &cobra.Command{
	Use:   "tcping",
	Short: "Tcping",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, _ := cmd.Flags().GetInt("count")
		interval, _ := cmd.Flags().GetDuration("interval")
		if interval < time.Millisecond*250 {
			return errors.New("interval too small, minimum 250ms")

		}
		tcpAddr, err := net.ResolveTCPAddr("tcp", args[0])
		if err != nil {
			return err
		}
		addr := tcpAddr.String()
		fmt.Printf("Tcpinging %s (%s) :\n", args[0], addr)
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
			conn, err := net.Dial("tcp", addr)
			t := time.Since(start)
			if err != nil {
				fmt.Printf("Unexpected error: %s\n", err)
			} else {
				conn.Close()
				fmt.Printf("Connected %s : time=%s\n", addr, t)
				success += 1
				sum += t
				if min == 0 || t < min {
					min = t
				}
				if t > max {
					max = t
				}
			}
			total += 1
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
	tcpingCmd.Flags().DurationP("interval", "i", time.Second, "time between sending each packet, minimum 400ms")
	tcpingCmd.Flags().IntP("count", "c", 3, "ping times, nonpositive number means infinity")
	rootCmd.AddCommand(tcpingCmd)
}

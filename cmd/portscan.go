package cmd

import (
	"awake/pkg/network"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/spf13/cobra"
)

func init() {
	var (
		timeout     time.Duration
		ports       []int
		portRange   string
		concurrency int
		verbose     bool
	)

	portScanCmd := &cobra.Command{
		Use:     "scan",
		Short:   "TCP port scanning",
		Example: "  awake scan 1.1.1.1 -p 80\n  awake scan 1.1.1.1 -r 80-443",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if u, err := url.Parse(args[0]); err == nil {
				if v := u.Hostname(); v != "" {
					args[0] = v
				}
			}
			if !network.IsDomain(args[0]) && !network.IsIP(args[0]) {
				return fmt.Errorf("invalid host: %s", args[0])
			}
			addrs, err := net.LookupHost(args[0])
			if err != nil {
				return err
			}
			ip := addrs[0]
			uniqueSet := make(map[int]struct{})
			for _, p := range ports {
				uniqueSet[p] = struct{}{}
			}
			if portRange != "" {
				ss := strings.Split(portRange, "-")
				if len(ss) != 2 {
					return errors.New("can't get port range")
				}
				start, _ := strconv.Atoi(ss[0])
				end, _ := strconv.Atoi(ss[1])
				for i := start; i <= end; i++ {
					uniqueSet[int(i)] = struct{}{}
				}
			}
			ports = ports[:0]
			for p := range uniqueSet {
				if p >= 0 && p <= 65535 {
					ports = append(ports, p)
				}
			}
			sort.Ints(ports)
			if len(ports) == 0 {
				return errors.New("no port to scan")
			}
			if concurrency > len(ports) {
				concurrency = len(ports)
			}
			fmt.Printf("Scanning %s (%s) with %d ports at %s\n", args[0], ip, len(ports), time.Now().Format(time.RFC3339))
			var wg sync.WaitGroup
			wg.Add(len(ports))
			var success int32
			fmt.Printf("%-5s  %-5s  Duration/Error\n", "Port", "Open")
			start := time.Now()
			pool, err := ants.NewPoolWithFunc(concurrency, func(i any) {
				defer wg.Done()
				port := i.(int)
				start := time.Now()
				conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, strconv.Itoa(port)), timeout)
				if err == nil {
					defer conn.Close()
					duration := time.Since(start)
					atomic.AddInt32(&success, 1)
					fmt.Printf("%-5d  true   %v\n", port, duration)
				} else if verbose {
					fmt.Printf("%-5d  false  %v\n", port, err)
				}
			})
			if err != nil {
				panic(err)
			}
			for _, p := range ports {
				if err := pool.Invoke(p); err != nil {
					panic(err)
				}
			}
			wg.Wait()
			fmt.Printf("\nTotal Time: %v  Num: %d  Open: %d  Closed: %d", time.Since(start), len(ports), success, len(ports)-int(success))
			return nil
		},
	}
	portScanCmd.Flags().StringVarP(&portRange, "port-range", "r", "", "port range, example: -r 1-100")
	portScanCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 128, "maximum concurrency")
	portScanCmd.Flags().IntSliceVarP(&ports, "port", "p", []int{}, "port to scan")
	portScanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show closed ports")
	portScanCmd.Flags().DurationVarP(&timeout, "timeout", "t", 6*time.Second, "connection timeout")
	rootCmd.AddCommand(portScanCmd)
}

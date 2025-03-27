package cmd

import (
	"awake/catbox"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload files",
	Long:  "Upload files to https://catbox.moe or https://litterbox.catbox.moe",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		concurrency, _ := cmd.Flags().GetInt("concurrency")
		concurrency = max(concurrency, 1)
		duration, _ := cmd.Flags().GetDuration("duration")
		proxy, _ := cmd.Flags().GetString("proxy")
		if !catbox.IsValidStorageDuration(duration) {
			logger.Fatalf("Invalid duration %s, must be 0, 12h, 24h, 72h", duration)
		}
		files := args
		if useGlob, _ := cmd.Flags().GetBool("glob"); useGlob {
			temp := []string{}
			for _, p := range files {
				matches, err := filepath.Glob(p)
				if err == nil {
					temp = append(temp, matches...)
				}
			}
			files = temp
		}
		if len(files) == 0 {
			logger.Warnln("No file to upload")
			return
		}
		if duration == 0 {
			fmt.Printf("A total of %d files will be stored permanently. Please enter the file count to continue: ", len(files))
			var inputNumber int
			fmt.Scan(&inputNumber)
			if inputNumber != len(files) {
				fmt.Println(logger.Yellow("Upload cancelled"))
				return
			}
		}
		if concurrency > len(files) {
			concurrency = len(files)
		}
		uploader := catbox.New()
		err := uploader.SetProxy(proxy)
		if err != nil {
			logger.Fatalf("Failed to set proxy: %s", err)
		}
		defer uploader.Close()
		var wg sync.WaitGroup
		wg.Add(len(files))
		ch := make(chan struct{}, concurrency)
		for _, f := range files {
			ch <- struct{}{}
			go func(file string) {
				defer func() {
					<-ch
					wg.Done()
				}()
				u, err := uploader.UploadFile(file, duration)
				if err == nil {
					fmt.Printf("%s    %s\n", file, logger.Green(u))
				} else {
					fmt.Printf("%s    %s\n", file, logger.Red(err.Error()))
				}
			}(f)
		}
		wg.Wait()
		close(ch)
	},
}

func init() {
	uploadCmd.Flags().IntP("concurrency", "c", 6, "maximum concurrency")
	uploadCmd.Flags().Bool("glob", false, "enable global syntax")
	uploadCmd.Flags().StringP("proxy", "p", "", "proxy url")
	uploadCmd.Flags().DurationP("duration", "d", time.Hour*12, "storage duration, 0 for permanent")
	rootCmd.AddCommand(uploadCmd)
}

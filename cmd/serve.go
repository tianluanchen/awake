package cmd

import (
	"awake/pkg"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	var addr string
	var cors bool
	var serveCmd = &cobra.Command{
		Use:     "serve",
		Short:   "Start static files server",
		Long:    "Start static files server, default directory is current directory",
		Args:    cobra.MaximumNArgs(1),
		Example: "  awake serve ./",
		Run: func(cmd *cobra.Command, args []string) {
			var dir string
			if len(args) == 0 {
				dir = "."
			} else {
				dir = args[0]
			}
			fileserver := http.FileServer(http.Dir(dir))
			srv := &http.Server{
				Addr: addr,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if cors {
						w.Header().Set("Access-Control-Allow-Origin", "*")
						w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
						w.Header().Set("Access-Control-Max-Age", "3600")
					}
					if r.Method == http.MethodOptions {
						w.WriteHeader(http.StatusNoContent)
						return
					}
					start := time.Now()
					fileserver.ServeHTTP(w, r)
					end := time.Now()
					duration := end.Sub(start)
					fmt.Println(end.Format("2006-01-02 15:04:05"), "|", duration, "|", r.Method, " ", r.RequestURI)
				}),
			}
			go func() {
				if resolved, err := pkg.ResolveListenAddr(addr); err == nil {
					for _, v := range resolved {
						if strings.HasPrefix(v, "127.0.0.1") || strings.HasPrefix(v, "::1") || strings.HasPrefix(v, "localhost") {
							fmt.Printf("Local:   http://%s\n", v)
						} else {
							fmt.Printf("Network: http://%s\n", v)
						}
					}
				}
				logger.Infoln("Server starting on", addr, "and serving", dir)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatalln(err)
				}
			}()
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt)
			<-quit
			if err := srv.Shutdown(context.Background()); err != nil {
				logger.Errorln(err)
			} else {
				logger.Warnln("Server exiting")
			}
		},
	}
	serveCmd.Flags().StringVarP(&addr, "addr", "a", "127.0.0.1:8080", "listen address")
	serveCmd.Flags().BoolVar(&cors, "cors", false, "enable CORS")
	rootCmd.AddCommand(serveCmd)

}

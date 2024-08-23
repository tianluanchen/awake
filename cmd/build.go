package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build binary file for golang project",
	Run: func(cmd *cobra.Command, args []string) {
		cocurrency, _ := cmd.Flags().GetInt("cocurrency")
		if cocurrency <= 0 {
			cocurrency = runtime.NumCPU()
		}
		output, _ := cmd.Flags().GetString("output")
		ldflags, _ := cmd.Flags().GetString("ldflags")
		trimpath, _ := cmd.Flags().GetBool("trimpath")
		list, err := getValidGolangTargets()
		if err != nil {
			logger.Fatalln(err)
		}
		if len(args) == 0 {
			args = append(args, runtime.GOOS+"/"+runtime.GOARCH)
		}
		set := make(map[string]bool)
		targets := make([][]string, 0, 16)
		for _, v := range args {
			s := strings.Split(v, "/")
			if s[0] == "" {
				s[0] = runtime.GOOS
			}
			if len(s) == 1 {
				s = append(s, runtime.GOARCH)
			} else if len(s) > 2 {
				logger.Fatalln("invalid target:", v)
			}
			if s[1] == "" {
				s[1] = runtime.GOARCH
			}
			t := strings.Join(s, "/")
			var exist bool
			for _, v := range list {
				if v == t {
					exist = true
					break
				}
			}
			if !exist {
				logger.Fatalln("unsupported target:", v)
			}
			if set[t] {
				continue
			}
			set[t] = true
			targets = append(targets, s)
		}
		if len(targets) == 0 {
			logger.Fatalln("no target specified")
		}
		isSingleTarget := len(targets) == 1
		modName, err := getGoModName()
		if err != nil {
			logger.Fatalln(err)
		}
		var outputFile string
		if output == "" && isSingleTarget {
			outputFile = modName
		} else {
			output = path.Clean(output)
			info, err := os.Stat(output)
			if err == nil {
				if info.IsDir() {
					if isSingleTarget {
						outputFile = fixBinaryFileName(path.Join(output, modName), targets[0][0])
					}
				} else {
					if isSingleTarget {
						outputFile = output
					}
				}
			} else if !os.IsNotExist(err) {
				logger.Fatalln(err)
			} else {
				dir := output
				if isSingleTarget {
					outputFile = output
					dir = path.Dir(output)
				}
				if dir != "." {
					logger.Warnln("mkdir", dir)
					if err := os.MkdirAll(dir, 0755); err != nil {
						logger.Fatalln(err)
					}
				}
			}
		}
		var wg sync.WaitGroup
		ch := make(chan struct{}, cocurrency)
		for _, s := range targets {
			goos, goarch := s[0], s[1]
			f := outputFile
			if f == "" {
				f = fixBinaryFileName(path.Join(output, fmt.Sprintf("%s_%s_%s", modName, goos, goarch)), goos)
			}
			fmt.Println("building for", goos+"/"+goarch, "===>", f)
			wg.Add(1)
			ch <- struct{}{}
			go func() {
				defer func() {
					wg.Done()
					<-ch
				}()
				args := []string{"build", "-o", f, "-ldflags", ldflags}
				if trimpath {
					args = append(args,
						"-trimpath",
					)
				}
				cmd := exec.Command("go", args...)
				cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
				b, err := cmd.CombinedOutput()
				if err == nil {
					fmt.Println(logger.Green("build " + f + " success"))
					return
				}
				fmt.Println(logger.Red("build " + f + " failed: " + string(b)))
			}()
		}
		wg.Wait()
		close(ch)
	},
}

func fixBinaryFileName(name string, target string) string {
	if strings.HasSuffix(name, ".exe") {
		return name
	}
	if strings.Contains(target, "windows") {
		return name + ".exe"
	}
	return name
}
func getGoModName() (string, error) {
	file, err := os.Open("go.mod")
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " \r\t")
		if strings.HasPrefix(line, "module") {
			v := strings.Trim(line[len("module"):], " \r\t")
			if name := path.Base(v); name == "." || name == "/" {
				return "", fmt.Errorf("invalid module name: %s", v)
			} else {
				return name, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", errors.New("cannot find module name")
}

func getValidGolangTargets() ([]string, error) {
	cmd := exec.Command("go", "tool", "dist", "list")
	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(b), "\n"), nil
}

func init() {
	buildCmd.Flags().IntP("cocurrency", "c", 6, "number of concurrent goroutines")
	buildCmd.Flags().StringP("output", "o", "", "output file or directory")
	buildCmd.Flags().String("ldflags", "-s -w", "ldflags")
	buildCmd.Flags().Bool("trimpath", false, "trim path")
	rootCmd.AddCommand(buildCmd)
}

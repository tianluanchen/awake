package cmd

import (
	"awake/pkg"
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
		targetList, _ := cmd.Flags().GetStringSlice("target")
		outputDir, _ := cmd.Flags().GetString("output")
		tags, _ := cmd.Flags().GetString("tags")
		ldflags, _ := cmd.Flags().GetString("ldflags")
		trimpath, _ := cmd.Flags().GetBool("trimpath")
		format, _ := cmd.Flags().GetString("format")
		env, _ := cmd.Flags().GetStringArray("env")
		validTargetList, err := getValidGolangTargets()
		if err != nil {
			logger.Fatalln(err)
		}
		if len(targetList) == 0 {
			targetList = append(targetList, runtime.GOOS+"/"+runtime.GOARCH)
		}
		set := make(map[string]bool)
		targets := make([]*buildTarget, 0, 16)
		for _, v := range targetList {
			s := strings.Split(v, "/")
			if s[0] == "" {
				s[0] = runtime.GOOS
			}
			if len(s) == 1 {
				s = append(s, runtime.GOARCH)
			} else if len(s) > 2 {
				logger.Fatalln("invalid target", v)
			}
			if s[1] == "" {
				s[1] = runtime.GOARCH
			}
			t := strings.Join(s, "/")
			var exist bool
			for _, v := range validTargetList {
				if v == t {
					exist = true
					break
				}
			}
			if !exist {
				logger.Fatalln("unsupported target", v)
			}
			if set[t] {
				continue
			}
			set[t] = true
			targets = append(targets, &buildTarget{
				GOOS:   s[0],
				GOARCH: s[1],
			})
		}
		if len(targets) == 0 {
			logger.Fatalln("no target specified")
		}
		modName, err := getGoModName()
		if err != nil {
			logger.Fatalln(err)
		}
		outputDir = path.Clean(outputDir)
		info, err := os.Stat(outputDir)
		if err == nil {
			if !info.IsDir() {
				logger.Fatalln(outputDir, "exists and is not a directory!")
			}
		} else if !os.IsNotExist(err) {
			logger.Fatalln(err)
		}
		for _, t := range targets {
			t.Output = path.Join(outputDir, t.AddExt(getNameWithFormat(format, modName, t)))
		}
		var wg sync.WaitGroup
		var failed bool
		ch := make(chan struct{}, cocurrency)
		for _, t := range targets {
			fmt.Println("building for", t.OSARCH(), "===>", t.Output)
			wg.Add(1)
			ch <- struct{}{}
			go func() {
				defer func() {
					wg.Done()
					<-ch
				}()
				cmdArgs := []string{"build", "-o", t.Output, "-ldflags", ldflags}
				if trimpath {
					cmdArgs = append(cmdArgs,
						"-trimpath",
					)
				}
				if tags != "" {
					cmdArgs = append(cmdArgs,
						"-tags", tags,
					)
				}
				cmdArgs = append(cmdArgs, args...)
				cmd := exec.Command("go", cmdArgs...)
				cmd.Env = append(os.Environ(), env...)
				cmd.Env = append(cmd.Env, "GOOS="+t.GOOS, "GOARCH="+t.GOARCH)
				b, err := cmd.CombinedOutput()
				if err == nil {
					var sizeStr string
					info, err := os.Stat(t.Output)
					if err != nil {
						sizeStr = err.Error()
					} else {
						sizeStr = pkg.FormatSize(info.Size())
					}
					fmt.Println(logger.Green(t.Output + "    " + sizeStr))
					return
				}
				failed = true
				fmt.Println(logger.Red("failed to build " + t.Output + ": " + string(b)))
			}()
		}
		wg.Wait()
		close(ch)
		if failed {
			os.Exit(1)
		}
	},
}

type buildTarget struct {
	GOOS   string
	GOARCH string
	Output string
}

func (t *buildTarget) OSARCH() string {
	return t.GOOS + "/" + t.GOARCH
}

// if GOOS is windows, add .exe
func (t *buildTarget) AddExt(name string) string {
	if strings.HasSuffix(name, ".exe") {
		return name
	}
	if t.GOOS == "windows" {
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

func getNameWithFormat(format, mod string, t *buildTarget) string {
	format = strings.ReplaceAll(format, "{{.MOD}}", mod)
	format = strings.ReplaceAll(format, "{{.OS}}", t.GOOS)
	format = strings.ReplaceAll(format, "{{.ARCH}}", t.GOARCH)
	return format
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
	buildCmd.Flags().StringP("output", "o", ".", "output directory")
	buildCmd.Flags().String("ldflags", "-s -w", "ldflags")
	buildCmd.Flags().String("tags", "", "tags")
	buildCmd.Flags().Bool("trimpath", false, "trim path")
	buildCmd.Flags().StringSlice("target", []string{}, "target os/arch, eg. linux/amd64, windows")
	buildCmd.Flags().StringArrayP("env", "e", []string{}, "set environment variables, eg. CGO_ENABLED=0")
	buildCmd.Flags().StringP("format", "f", "{{.MOD}}_{{.OS}}_{{.ARCH}}", "basic name format")
	rootCmd.AddCommand(buildCmd)
}

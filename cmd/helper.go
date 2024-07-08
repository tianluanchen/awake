package cmd

import (
	"awake/pkg"
	"awake/pkg/network"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

var logger = pkg.NewLogger()

func newBar(max int64, options ...progressbar.Option) *progressbar.ProgressBar {
	options = append([]progressbar.Option{
		progressbar.OptionUseIECUnits(true),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionUseANSICodes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionShowBytes(true),
		progressbar.OptionThrottle(time.Millisecond * 100),
		progressbar.OptionShowCount(),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	}, options...)
	return progressbar.NewOptions64(max, options...)
}

func parseResolveFlag(hostname string, resolveArr ...string) (map[string]string, error) {
	resolveHostMap := make(map[string]string)
	for _, v := range resolveArr {
		ss := strings.Split(v, ":")
		err := errors.New("can't resolve host: " + v)
		if len(ss) == 1 {
			if network.IsIP(ss[0]) {
				resolveHostMap[hostname] = ss[0]
			} else {
				return nil, err
			}
		} else if len(ss) == 2 {
			if (network.IsDomain(ss[0]) || ss[0] == "*") && network.IsIP(ss[1]) {
				resolveHostMap[ss[0]] = ss[1]
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return resolveHostMap, nil
}

func parseHeader(headerArr ...string) http.Header {
	header := make(http.Header)
	for _, v := range headerArr {
		index := strings.Index(v, ":")
		if index > -1 {
			k := strings.Trim(v[0:index], "\r\n\t ")
			v := strings.Trim(v[index+1:], "\r\n\t ")
			if len(k) > 0 {
				header.Add(k, v)
			}
		}
	}
	return header
}

package catbox

import (
	"awake/pkg"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// https://catbox.moe/
//
// https://litterbox.catbox.moe/
type Uploader struct {
	client *http.Client
	header http.Header
}

func New() *Uploader {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	up := &Uploader{
		client: &http.Client{
			Transport: transport,
		},
		header: http.Header{
			"User-Agent": []string{"Mozilla/5.0 (X11; Linux i686; rv:136.0) Gecko/20100101 Firefox/136.0"},
			"Accept":     []string{"*/*"},
		},
	}
	return up
}

func (up *Uploader) Close() error {
	up.client.CloseIdleConnections()
	return nil
}

// storageDuration: 0, 1h, 12h, 24h, 72h
//
// If set to 0, data is stored permanently.
func (up *Uploader) Upload(r io.Reader, name string, mimeType string, storageDuration time.Duration) (string, error) {
	var apiURL string
	switch storageDuration {
	case 0:
		apiURL = "https://catbox.moe/user/api.php"
	case time.Hour, time.Hour * 12, time.Hour * 24, time.Hour * 72:
		apiURL = "https://litterbox.catbox.moe/resources/internals/api.php"
	default:
		return "", fmt.Errorf("unsupported storage duration: %s", storageDuration)
	}
	req, err := http.NewRequest(http.MethodPost, apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header = up.header
	fd := pkg.NewFormData()
	fd.Add(&pkg.FormFile{
		Name:     "fileToUpload",
		FileName: name,
		MimeType: mimeType,
		Reader:   r,
	})
	fd.Add(&pkg.FormFile{
		Name:  "reqtype",
		Value: "fileupload",
	})
	if storageDuration != 0 {
		fd.Add(&pkg.FormFile{
			Name:  "time",
			Value: strconv.Itoa(int(storageDuration/time.Hour)) + "h",
		})
	}
	fd.AttachToRequest(req)
	resp, err := up.client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		b := make([]byte, 1024)
		n, err := io.ReadFull(resp.Body, b)
		var s string
		if err == io.ErrUnexpectedEOF {
			s = string(b[:n])
		} else {
			s = string(b)
		}
		return "", fmt.Errorf("%s : %s...", resp.Status, s)
	}
	b := make([]byte, 1024)
	n, err := io.ReadAtLeast(resp.Body, b, len(b))
	if err == nil || err == io.ErrUnexpectedEOF {
		return string(b[:n]), nil
	}
	return "", err
}

// storageDuration: 0, 1h, 12h, 24h, 72h
//
// If set to 0, data is stored permanently.
func (up *Uploader) UploadFile(file string, storageDuration time.Duration) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory", file)
	}
	fileSize := info.Size()
	maxSize := MaxUploadSize(storageDuration)
	if fileSize > maxSize {
		return "", fmt.Errorf("file size exceeds limit: %s > %s", pkg.FormatSize(fileSize), pkg.FormatSize(maxSize))
	}
	ext := filepath.Ext(file)
	if ext != "" && IsUnsupportedExtension(ext) {
		ext = ".to-" + ext[1:]
	}
	return up.Upload(f, RandomName(6)+ext, "application/octet-stream", storageDuration)
}

func (up *Uploader) SetProxy(proxy string) error {
	if proxy == "" {
		up.client.Transport.(*http.Transport).Proxy = http.ProxyFromEnvironment
		return nil
	}
	if !strings.Contains(proxy, "://") {
		proxy = "http://" + proxy
	}
	u, err := url.Parse(proxy)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "socks5" {
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
	up.client.Transport.(*http.Transport).Proxy = http.ProxyURL(u)
	return nil
}

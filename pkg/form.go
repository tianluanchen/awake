package pkg

import (
	"crypto/rand"
	"io"
	"net/http"
	"strings"
)

type FormFile struct {
	Name  string
	Value string

	FileName string
	MimeType string
	Reader   io.Reader
}

type FormData struct {
	files    []*FormFile
	boundary string
}

func NewFormData() *FormData {
	b := make([]byte, 30)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(err)
	}
	for i := range len(b) {
		b[i] = 48 + b[i]%10
	}
	return &FormData{
		boundary: "---------------------------" + string(b),
	}
}

func (fd *FormData) ContentType() string {
	return "multipart/form-data; boundary=" + fd.boundary
}

func (fd *FormData) Add(f *FormFile) {
	fd.files = append(fd.files, f)
}

func (fd *FormData) AsReader() io.Reader {
	var rs []io.Reader
	for _, f := range fd.files {
		if f.Reader == nil {
			rs = append(rs, strings.NewReader("--"+fd.boundary+"\r\nContent-Disposition: form-data; name=\""+f.Name+"\"\r\n\r\n"+f.Value+"\r\n"))
		} else {
			rs = append(rs, strings.NewReader("--"+fd.boundary+"\r\nContent-Disposition: form-data; name=\""+f.Name+"\"; filename=\""+f.FileName+"\"\r\nContent-Type: "+f.MimeType+"\r\n\r\n"))
			rs = append(rs, f.Reader)
			rs = append(rs, strings.NewReader("\r\n"))
		}
	}
	rs = append(rs, strings.NewReader("--"+fd.boundary+"--\r\n"))
	return io.MultiReader(rs...)
}

// Excluding Reader size
func (fd *FormData) Size() int {
	var size int
	for _, f := range fd.files {
		if f.Reader == nil {
			size += 2 + len(fd.boundary) + 40 + len(f.Name) + 5 + len(f.Value) + 2
		} else {
			size += 2 + len(fd.boundary) + 40 + len(f.Name) + 13 + len(f.FileName) + 17 + len(f.MimeType) + 4
			size += 2
		}
	}
	size += 2 + len(fd.boundary) + 4
	return size
}

func (fd *FormData) AttachToRequest(req *http.Request) {
	req.Header.Set("Content-Type", fd.ContentType())
	req.ContentLength = -1
	req.Body = io.NopCloser(fd.AsReader())
}

package cmd

import (
	"awake/pkg"
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

var logger = pkg.NewLogger()

// dialWithProxy establishes a TCP connection to the target address through an HTTP or SOCKS5 proxy.
// proxyURL: URL of the proxy (e.g., "http://proxy.example.com:8080" or "socks5://proxy.example.com:1080").
// targetAddr: Address of the target in the form of "host:port".
func dialTCPWithProxy(proxyURL string, targetAddr string) (net.Conn, error) {
	if proxyURL == "" || targetAddr == "" {
		return nil, errors.New("proxyURL and targetAddr cannot be empty")
	}
	if !strings.Contains(proxyURL, "://") {
		proxyURL = "http://" + proxyURL
	}
	// Parse the proxy URL
	proxyURI, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}
	switch proxyURI.Scheme {
	case "http", "https":
		if proxyURI.Port() == "" {
			port := "80"
			if proxyURI.Scheme == "https" {
				port = "443"
			}
			proxyURI.Host = net.JoinHostPort(proxyURI.Host, port)
		}
		conn, err := net.Dial("tcp", proxyURI.Host)
		if err != nil {
			return nil, err
		}
		if proxyURI.Scheme == "https" {
			conn = tls.Client(conn, &tls.Config{
				ServerName: proxyURI.Hostname(),
				NextProtos: []string{"http/1.1"},
			})
		}
		req := &http.Request{
			Method:     http.MethodConnect,
			Host:       targetAddr,
			RequestURI: targetAddr,
			URL:        proxyURI,
		}
		if user := proxyURI.User; user != nil {
			pwd, _ := user.Password()
			req.Header = http.Header{
				"Proxy-Authorization": {"Basic " + base64.StdEncoding.EncodeToString([]byte(user.Username()+":"+pwd))},
			}
		}
		err = req.Write(conn)
		if err != nil {
			return nil, err
		}
		bufR := bufio.NewReader(conn)
		resp, err := http.ReadResponse(bufR, req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(resp.Status)
		}
		return &readerConn{reader: bufR, Conn: conn}, nil
	case "socks5":
		if proxyURI.Port() == "" {
			proxyURI.Host = net.JoinHostPort(proxyURI.Host, "1080")
		}
		var auth *proxy.Auth
		if user := proxyURI.User; user != nil {
			pwd, _ := user.Password()
			auth = &proxy.Auth{
				User:     user.Username(),
				Password: pwd,
			}
		}
		dialer, err := proxy.SOCKS5("tcp", proxyURI.Host, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}
		return dialer.Dial("tcp", targetAddr)
	default:
		return nil, errors.New("unsupported proxy scheme")
	}
}

type readerConn struct {
	reader *bufio.Reader
	net.Conn
}

func (r *readerConn) Read(b []byte) (int, error) {
	return r.reader.Read(b)
}

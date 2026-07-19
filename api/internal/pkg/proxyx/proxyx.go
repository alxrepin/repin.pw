package proxyx

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

func init() {
	proxy.RegisterDialerType("http", newConnectDialer)
	proxy.RegisterDialerType("https", newConnectDialer)
}

func Dialer(rawURL string) (DialFunc, error) {
	if rawURL == "" {
		return nil, nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse proxy url: %w", err)
	}

	dialer, err := proxy.FromURL(u, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("proxy dialer: %w", err)
	}

	ctxDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		return nil, fmt.Errorf("proxy scheme %q does not support context dialing", u.Scheme)
	}

	return ctxDialer.DialContext, nil
}

func HTTPClient(rawURL string, timeout time.Duration) (*http.Client, error) {
	if rawURL == "" {
		return &http.Client{Timeout: timeout}, nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse proxy url: %w", err)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()

	if u.Scheme == "http" || u.Scheme == "https" {
		transport.Proxy = http.ProxyURL(u)
	} else {
		dial, err := Dialer(rawURL)
		if err != nil {
			return nil, err
		}

		transport.Proxy = nil
		transport.DialContext = dial
	}

	return &http.Client{Timeout: timeout, Transport: transport}, nil
}

type connectDialer struct {
	address string
	auth    *proxy.Auth
	tls     bool
	forward proxy.Dialer
}

func newConnectDialer(u *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	d := &connectDialer{
		address: u.Host,
		tls:     u.Scheme == "https",
		forward: forward,
	}

	if u.Port() == "" {
		port := "80"
		if d.tls {
			port = "443"
		}

		d.address = net.JoinHostPort(u.Hostname(), port)
	}

	if u.User != nil {
		password, _ := u.User.Password()
		d.auth = &proxy.Auth{User: u.User.Username(), Password: password}
	}

	return d, nil
}

func (d *connectDialer) Dial(network, addr string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}

func (d *connectDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("http proxy: unsupported network %q", network)
	}

	conn, err := d.dialProxy(ctx, network)
	if err != nil {
		return nil, fmt.Errorf("http proxy: connect to proxy: %w", err)
	}

	if err := d.handshake(ctx, conn, addr); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return conn, nil
}

func (d *connectDialer) dialProxy(ctx context.Context, network string) (net.Conn, error) {
	var (
		conn net.Conn
		err  error
	)

	if dialer, ok := d.forward.(proxy.ContextDialer); ok {
		conn, err = dialer.DialContext(ctx, network, d.address)
	} else {
		conn, err = d.forward.Dial(network, d.address)
	}

	if err != nil || !d.tls {
		return conn, err
	}

	host, _, err := net.SplitHostPort(d.address)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return tls.Client(conn, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}), nil
}

func basicAuth(user, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + password))
}

func (d *connectDialer) handshake(ctx context.Context, conn net.Conn, addr string) error {
	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return fmt.Errorf("http proxy: set deadline: %w", err)
		}
	}

	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: make(http.Header),
	}

	if d.auth != nil {
		req.Header.Set("Proxy-Authorization", "Basic "+basicAuth(d.auth.User, d.auth.Password))
	}

	if err := req.Write(conn); err != nil {
		return fmt.Errorf("http proxy: write CONNECT: %w", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return fmt.Errorf("http proxy: read CONNECT response: %w", err)
	}

	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http proxy: CONNECT %s: %s", addr, resp.Status)
	}

	return conn.SetDeadline(time.Time{})
}

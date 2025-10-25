package reloader

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DevJHansen/go-reload/config"
)

//go:embed socket.html
var socketFile string

type Proxy struct {
	proxyServer *http.Server
}

func Start(c *config.Config) (Proxy, error) {
	targetUrl, err := url.Parse(fmt.Sprintf("http://localhost:%d", c.AppPort))
	if err != nil {
		return Proxy{}, fmt.Errorf("failed to parse target URL: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	proxy.ModifyResponse = func(res *http.Response) error {
		contentType := res.Header.Get("Content-Type")

		if strings.Contains(contentType, "text/html") {
			body, err := io.ReadAll(res.Body)

			if err != nil {
				return fmt.Errorf("Error reading body: %v", err)
			}
			res.Body.Close()

			stringyfyBody := string(body)

			socketScript := string(socketFile)
			modifiedBody := strings.Replace(stringyfyBody, "</body>", socketScript+"</body>", 1)
			parseBody := []byte(modifiedBody)

			res.Body = io.NopCloser(bytes.NewReader(parseBody))
			res.ContentLength = int64(len(modifiedBody))
			res.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))
		}

		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	proxyServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.ProxyPort),
		Handler: mux,
	}

	fmt.Printf("Proxy running on :%d, forwarding to :%d\n", c.ProxyPort, c.AppPort)

	go func() {
		if err := proxyServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Proxy server error: %v", err)
		}
	}()

	return Proxy{proxyServer: proxyServer}, nil
}

func (proxy *Proxy) Stop() error {
	if proxy == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := proxy.proxyServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("Falied to shutdown proxy: %w", err)
	}

	return nil
}

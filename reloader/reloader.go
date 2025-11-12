package reloader

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DevJHansen/go-reload/config"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

//go:embed socket.html
var socketFile string

type Proxy struct {
	proxyServer *http.Server
	clients     map[*websocket.Conn]bool
	clientsMu   sync.RWMutex
	upgrader    websocket.Upgrader
}

func Start(c *config.Config) (*Proxy, error) {
	p := &Proxy{
		clients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	targetUrl, err := url.Parse(fmt.Sprintf("http://localhost:%d", c.AppPort))
	if err != nil {
		return &Proxy{}, fmt.Errorf("failed to parse target URL: %w", err)
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

	mux.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		conn, err := p.upgrader.Upgrade(w, r, nil)

		if err != nil {
			color.Red("Error upgrading to web socket: %v", err)
			return
		}

		defer func() {
			p.clientsMu.Lock()
			delete(p.clients, conn)
			p.clientsMu.Unlock()
			conn.Close()
		}()

		p.clientsMu.Lock()
		p.clients[conn] = true
		p.clientsMu.Unlock()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}

	})

	p.proxyServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", c.ProxyPort),
		Handler: mux,
	}

	color.Cyan("Proxy running on :%d, forwarding to :%d", c.ProxyPort, c.AppPort)

	go func() {
		if err := p.proxyServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			color.Red("Proxy server error: %v", err)
		}
	}()

	return p, nil
}

func (proxy *Proxy) Stop() error {
	if proxy == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := proxy.proxyServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("Failed to shutdown proxy: %w", err)
	}

	return nil
}

func (p *Proxy) Broadcast(message string) {
	p.clientsMu.RLock()
	defer p.clientsMu.RUnlock()

	for client := range p.clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			color.Red("Error broadcasting to client: %v", err)
		}
	}
}

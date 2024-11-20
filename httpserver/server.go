package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/easysy/proton/coder"
)

type Formatter interface {
	coder.Coder
	WriteResponse(ctx context.Context, w http.ResponseWriter, statusCode int, v any)
}

// NewFormatter returns a new Formatter.
func NewFormatter(coder coder.Coder) Formatter {
	return &protoFormatter{Coder: coder}
}

type protoFormatter struct {
	coder.Coder
}

// WriteResponse encodes the value pointed to by v and writes it and statusCode to the stream.
func (f *protoFormatter) WriteResponse(ctx context.Context, w http.ResponseWriter, statusCode int, v any) {
	if v == nil {
		w.WriteHeader(statusCode)
		return
	}

	if w.Header().Get(coder.ContentType) == "" && f.ContentType() != "" {
		w.Header().Set(coder.ContentType, f.ContentType())
	}
	w.WriteHeader(statusCode)
	if err := f.Encode(ctx, w, v); err != nil {
		slog.ErrorContext(ctx, "encode response", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// Controller is a wrapper around *http.Server to control the server.
//
//	Server — *http.Server, which will be managed.
//	GracefulTimeout — time that is given to the server to shut down gracefully.
type Controller struct {
	Server          *http.Server
	GracefulTimeout time.Duration

	isRan   atomic.Bool
	restart atomic.Bool

	sigint  chan os.Signal
	mu      sync.Mutex
	onStart func(*http.Server)
}

// OnStart registers a callback function that is executed every time the controller starts or restarts.
// The provided function `f` receives a pointer to the HTTP server managed by the controller, allowing
// the user to perform custom initialization or configuration tasks at startup.
func (c *Controller) OnStart(f func(*http.Server)) {
	c.mu.Lock()
	c.onStart = f
	c.mu.Unlock()
}

// Start starts the *http.Server.
// If *tls.Config on the server is non nil, the server listens and serves using tls.
func (c *Controller) Start() (err error) {
	for {
		if err = c.start(); errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		if !c.restart.Load() {
			slog.Info("HTTP server is shutdown")
			return
		} else if err != nil {
			slog.Error(err.Error())
		}

		slog.Info("HTTP server is restarting")

		c.clone()
		c.restart.Store(false)
	}
}

// Restart restarts the server if necessary.
// For changes to the following parameters to take effect:
//
//	Addr; TLSConfig; TLSNextProto; ConnState; BaseContext; ConnContext,
//
// a server restart is required.
// Other parameters can be changed without restarting the server.
// If the server is not running, the function will be skipped.
func (c *Controller) Restart() {
	if !c.isRan.Load() {
		return
	}

	c.restart.Store(true)

	c.sigint <- syscall.SIGINT
}

// Shutdown gracefully shuts down the server.
func (c *Controller) Shutdown() {
	ctx, cancelWithTimeout := context.WithTimeout(context.Background(), c.GracefulTimeout)
	defer cancelWithTimeout()

	if err := c.Server.Shutdown(ctx); err != nil {
		slog.Error(fmt.Sprintf("HTTP server shutdown: %s", err))
	}
}

func (c *Controller) start() error {
	c.isRan.Store(true)

	c.sigint = make(chan os.Signal, 1)
	signal.Notify(c.sigint, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		c.isRan.Store(false)
		signal.Stop(c.sigint)
		close(c.sigint)
	}()

	go func() {
		<-c.sigint
		c.Shutdown()
	}()

	c.mu.Lock()
	if c.onStart != nil {
		c.onStart(c.Server)
	}
	c.mu.Unlock()

	secure := c.Server.TLSConfig != nil
	slog.Info("HTTP server serving", "secure", secure, "address", c.Server.Addr)

	if secure {
		err := c.Server.ListenAndServeTLS("", "")
		return fmt.Errorf("HTTP server ListenAndServeTLS: %w", err)
	} else {
		err := c.Server.ListenAndServe()
		return fmt.Errorf("HTTP server ListenAndServe: %w", err)
	}
}

// clone clones the server before restarting, since it is impossible to start a stopped server.
func (c *Controller) clone() {
	if c.Server.TLSConfig != nil && len(c.Server.TLSConfig.Certificates) == 0 {
		c.Server.TLSConfig = nil
	}

	c.Server = &http.Server{
		Addr:                         c.Server.Addr, // need to restart
		Handler:                      c.Server.Handler,
		DisableGeneralOptionsHandler: c.Server.DisableGeneralOptionsHandler,
		TLSConfig:                    c.Server.TLSConfig, // need to restart
		ReadTimeout:                  c.Server.ReadTimeout,
		ReadHeaderTimeout:            c.Server.ReadHeaderTimeout,
		WriteTimeout:                 c.Server.WriteTimeout,
		IdleTimeout:                  c.Server.IdleTimeout,
		MaxHeaderBytes:               c.Server.MaxHeaderBytes,
		TLSNextProto:                 c.Server.TLSNextProto, // need to restart
		ConnState:                    c.Server.ConnState,    // need to restart
		ErrorLog:                     c.Server.ErrorLog,
		BaseContext:                  c.Server.BaseContext, // need to restart
		ConnContext:                  c.Server.ConnContext, // need to restart
	}
}

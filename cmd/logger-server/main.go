package main

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"ella.to/httputil"
	"ella.to/sse"

	"ella.to/logger/www"
)

var (
	version     = "v0.0.3"
	defaultAddr = "localhost:2022"
	env         = strings.ToLower(getStringEnv("LOGGER_ENV", "prod"))
	isDev       = env == "dev"
)

type Stream struct {
	rwlock    sync.RWMutex
	conns     map[int64]sse.Pusher
	idCounter int64
}

func (s *Stream) Add(pusher sse.Pusher) func() {
	s.rwlock.Lock()
	s.idCounter++
	id := s.idCounter
	s.conns[id] = pusher
	s.rwlock.Unlock()

	return func() {
		s.rwlock.Lock()
		delete(s.conns, id)
		s.rwlock.Unlock()
	}
}

func (s *Stream) Broadcast(ctx context.Context, msg string) error {
	s.rwlock.RLock()
	defer s.rwlock.RUnlock()

	for _, pusher := range s.conns {
		err := pusher.Push(ctx, "message", msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	fmt.Printf(`
██╗░░░░░░█████╗░░██████╗░░██████╗░███████╗██████╗░
██║░░░░░██╔══██╗██╔════╝░██╔════╝░██╔════╝██╔══██╗
██║░░░░░██║░░██║██║░░██╗░██║░░██╗░█████╗░░██████╔╝
██║░░░░░██║░░██║██║░░╚██╗██║░░╚██╗██╔══╝░░██╔══██╗
███████╗╚█████╔╝╚██████╔╝╚██████╔╝███████╗██║░░██║
╚══════╝░╚════╝░░╚═════╝░░╚═════╝░╚══════╝╚═╝░░╚═╝ %s
`, version)

	ctx := context.Background()

	if len(os.Args) > 1 {
		defaultAddr = os.Args[1]
	}

	fmt.Println("Running on ", defaultAddr)

	stream := &Stream{
		conns: make(map[int64]sse.Pusher),
	}

	api := http.NewServeMux()

	api.Handle(
		"POST /logs",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			defer r.Body.Close()

			scanner := bufio.NewScanner(r.Body)
			scanner.Split(bufio.ScanLines)

			for scanner.Scan() {
				line := scanner.Text()
				if err := stream.Broadcast(ctx, line); err != nil {
					http.Error(w, "failed to broadcast message", http.StatusInternalServerError)
					slog.ErrorContext(ctx, "failed to broadcast message", "err", err)
					return
				}
			}

			if err := scanner.Err(); err != nil {
				http.Error(w, "failed to scan request body", http.StatusInternalServerError)
				slog.ErrorContext(ctx, "failed to scan request body", "err", err)
				return
			}

			w.WriteHeader(http.StatusOK)
		}),
	)

	api.Handle(
		"GET /logs",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			pusher, err := sse.CreatePusher(w)
			if err != nil {
				http.Error(w, "failed to create pusher", http.StatusInternalServerError)
				slog.ErrorContext(ctx, "failed to create pusher", "err", err)
				return
			}

			defer pusher.Done(ctx)

			unsubscribe := stream.Add(pusher)
			defer unsubscribe()

			<-ctx.Done()
		}),
	)

	//

	srv := &http.Server{
		Addr: defaultAddr,
	}

	if isDev {
		mux := http.NewServeMux()
		httputil.DevProxy(mux, api, isDev, "http://localhost:5173", []string{
			"/logs",
		})
		srv.Handler = mux
	} else {
		contentStatic, _ := fs.Sub(www.Files, "dist")
		fileServer := httputil.ServeFile(contentStatic)
		srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/logs") {
				api.ServeHTTP(w, r)
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	go func() {
		slog.InfoContext(ctx, "starting server", "addr", defaultAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.ErrorContext(ctx, "failed to listen and serve", "err", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	if err := srv.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to shutdown server", "err", err)
		os.Exit(1)
	}
}

func getStringEnv(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	return value
}

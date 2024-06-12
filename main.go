package main

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

func main() {
	increaseCounter := Counter(NewDB())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		counter, err := increaseCounter(r.Context())
		if err != nil {
			slog.Error("increase counter", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		_, err = w.Write(CounterResponse(counter))
		if err != nil {
			slog.Error("send counter response", slog.Any("error", err))

			return
		}
	})

	l, err := net.Listen("tcp", ":3000") //nolint:gosec
	if err != nil {
		slog.Error("open TCP port for listenening", slog.Any("error", err))

		return
	}

	slog.Info("ready")

	server := &http.Server{
		ReadHeaderTimeout: time.Second,
	}

	err = server.Serve(l)
	if err != nil {
		slog.Error("run counter server", slog.Any("error", err))

		return
	}
}

func CounterResponse(counter int) []byte {
	return []byte("counter: " + strconv.Itoa(counter))
}

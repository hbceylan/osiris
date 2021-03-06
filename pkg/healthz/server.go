package healthz

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"k8s.io/klog"
)

func RunServer(ctx context.Context, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", HandleHealthCheckRequest)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	doneCh := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done(): // Context was canceled or expired
			klog.Info("Healthz server is shutting down")
			// Allow up to five seconds for requests in progress to be completed
			shutdownCtx, cancel := context.WithTimeout(
				context.Background(),
				time.Second*5,
			)
			defer cancel()
			srv.Shutdown(shutdownCtx) // nolint: errcheck
		case <-doneCh: // The server shut down on its own, perhaps due to error
		}
	}()

	klog.Infof("Healthz server is listening on %s", srv.Addr)
	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		klog.Errorf("Healthz server error: %s", err)
		err = nil
	}
	close(doneCh)
}

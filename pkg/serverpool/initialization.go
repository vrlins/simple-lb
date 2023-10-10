package serverpool

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/vrlins/simple-lb/pkg/backend"
	"github.com/vrlins/simple-lb/pkg/utils"
)

func InitializeServerPool(serverList string, sp *ServerPool) {
	tokens := strings.Split(serverList, ",")
	for _, tok := range tokens {
		serverUrl, err := url.Parse(tok)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			log.Printf("[%s] %s\n", serverUrl.Host, e.Error())
			retries := utils.GetRetryFromContext(r)
			if retries < utils.MaxRetry {
				select {
				case <-time.After(utils.RetryWaitDuration):
					ctx := context.WithValue(r.Context(), utils.Retry, retries+1)
					proxy.ServeHTTP(w, r.WithContext(ctx))
				}
				return
			}

			sp.MarkBackendStatus(serverUrl, false)

			attempts := utils.GetAttemptsFromContext(r)
			log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
			ctx := context.WithValue(r.Context(), utils.Attempts, attempts+1)
			sp.GetNextPeer().ReverseProxy.ServeHTTP(w, r.WithContext(ctx)) // Use the next available peer for retry
		}

		sp.AddBackend(&backend.Backend{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		})
		log.Printf("Configured server: %s\n", serverUrl)
	}
}

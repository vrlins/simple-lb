package serverpool

import (
	"log"
	"net"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/vrlins/simple-lb/pkg/backend"
	"github.com/vrlins/simple-lb/pkg/utils"
)

type ServerPool struct {
	Backends []*backend.Backend
	Current  uint64
}

func (s *ServerPool) AddBackend(bk *backend.Backend) {
	s.Backends = append(s.Backends, bk)
}

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.Current, 1) % uint64(len(s.Backends)))
}

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.Backends {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

func (s *ServerPool) GetNextPeer() *backend.Backend {
	next := s.NextIndex()
	l := len(s.Backends) + next
	for i := next; i < l; i++ {
		idx := i % len(s.Backends)
		if s.Backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.Current, uint64(idx))
			}
			return s.Backends[idx]
		}
	}
	return nil
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.Backends {
		status := "up"
		alive := s.isBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

func (s *ServerPool) isBackendAlive(u *url.URL) bool {
	conn, err := net.DialTimeout("tcp", u.Host, utils.BackendConnectionTimeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	defer conn.Close()
	return true
}

func HealthCheckRoutine(sp *ServerPool) {
	ticker := time.NewTicker(utils.HealthCheckInterval)
	for {
		select {
		case <-ticker.C:
			sp.HealthCheck()
		}
	}
}

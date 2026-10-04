package graceful

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

// 临时调试用例：复现 Shutdown 超时，并打印每条连接的 ConnState 轨迹。
func TestDebugLingeringConn(t *testing.T) {
	for attempt := 0; attempt < 40; attempt++ {
		var mu sync.Mutex
		var trace []string
		record := func(c net.Conn, st http.ConnState) {
			mu.Lock()
			trace = append(trace, fmt.Sprintf("%s->%v", c.RemoteAddr(), st))
			mu.Unlock()
		}

		started := make(chan struct{})
		mux := http.NewServeMux()
		mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
			close(started)
			select {
			case <-time.After(300 * time.Millisecond):
				fmt.Fprint(w, "done")
			case <-r.Context().Done():
				fmt.Fprint(w, "interrupted")
			}
		})

		lc := New(Config{Handler: mux, Logger: quietLogger(), DrainTimeout: 2 * time.Second})
		lc.srv.ConnState = record

		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		errCh := make(chan error, 1)
		go func() { errCh <- lc.Serve(ctx, ln) }()
		addr := "http://" + ln.Addr().String()

		// 复刻 start() 里的 waitReady
		wc := &http.Client{Timeout: time.Second}
		for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); {
			resp, err := wc.Get(addr + "/readyz")
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
		}

		go func() {
			c := &http.Client{Timeout: 3 * time.Second}
			resp, err := c.Get(addr + "/slow")
			if err == nil {
				io.ReadAll(resp.Body)
				resp.Body.Close()
			}
		}()
		<-started

		begin := time.Now()
		err = lc.Shutdown(context.Background())
		if err != nil {
			mu.Lock()
			t.Logf("attempt %d: Shutdown 等了 %v，err=%v", attempt, time.Since(begin).Round(time.Millisecond), err)
			t.Logf("ConnState 轨迹: %v", trace)
			mu.Unlock()
			cancel()
			<-errCh
			t.FailNow()
		}
		cancel()
		<-errCh
	}
}

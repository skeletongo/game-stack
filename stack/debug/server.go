package debug

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dobyte/due/v2/log"
)

// Server 是 debug HTTP 服务。
type Server struct {
	addr string
	mux  *http.ServeMux
}

// NewServer 创建 debug HTTP 服务。
func NewServer(addr string) *Server {
	s := &Server{
		addr: addr,
		mux:  http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// Start 启动 HTTP 服务（阻塞）。
func (s *Server) Start() error {
	log.Infof("[debug] server starting on %s", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}

// StartAsync 在后台 goroutine 启动服务。
func (s *Server) StartAsync() {
	go func() {
		if err := s.Start(); err != nil {
			log.Errorf("[debug] server error: %v", err)
		}
	}()
}

// registerRoutes 注册所有端点。
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/debug/modules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		handleDiscover(w, r)
	})

	s.mux.HandleFunc("/debug/module/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/debug/module/")
		if name == "" {
			writeError(w, http.StatusBadRequest, "module name required")
			return
		}
		handleModuleInfo(w, r, name)
	})

	s.mux.HandleFunc("/debug/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		handleQuery(w, r)
	})

	s.mux.HandleFunc("/debug/command", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		handleCommand(w, r)
	})

	s.mux.HandleFunc("/debug/patch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		handlePatch(w, r)
	})

	// Swagger
	s.mux.HandleFunc("/debug/swagger.json", handleSwaggerJSON)
	s.mux.HandleFunc("/debug/swagger/", handleSwaggerUI)
}

// writeJSON 写 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError 写错误响应。
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

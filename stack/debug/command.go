package debug

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/ddd"
)

// commandRequest 执行命令的请求体。
type commandRequest struct {
	Module string         `json:"module" example:"player"`
	Cmd    string         `json:"cmd"    example:"add_gold"`
	Params map[string]any `json:"params"`
}

// @Summary      执行命令
// @Description  通过 CommandBus 执行已注册的命令。命令处理器自动发布领域事件。
// @Description  命令列表可通过 GET /debug/module/{name} 查询。
// @Tags         command
// @Accept       json
// @Produce      json
// @Param        request  body  commandRequest  true  "命令请求"
// @Success      200      {object}  object  "命令执行结果"
// @Failure      400      {object}  map[string]string  "参数错误"
// @Failure      404      {object}  map[string]string  "模块或命令不存在"
// @Router       /debug/command [post]
func handleCommand(w http.ResponseWriter, r *http.Request) {
	var req commandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	if req.Module == "" || req.Cmd == "" {
		writeError(w, http.StatusBadRequest, "module and cmd are required")
		return
	}

	m, ok := get(req.Module)
	if !ok {
		writeError(w, http.StatusNotFound, "module not found: "+req.Module)
		return
	}
	if m.CmdBus == nil {
		writeError(w, http.StatusBadRequest, "module has no command bus: "+req.Module)
		return
	}

	cmd, ok := m.CmdBus.NewCommand(req.Cmd)
	if !ok {
		writeError(w, http.StatusNotFound, "command not found: "+req.Cmd)
		return
	}

	if req.Params != nil {
		b, _ := json.Marshal(req.Params)
		if err := json.Unmarshal(b, cmd); err != nil {
			writeError(w, http.StatusBadRequest, "invalid params: "+err.Error())
			return
		}
	}

	ret, err := m.CmdBus.Dispatch(context.Background(), cmd.(ddd.Command))
	if err != nil {
		log.Errorf("[debug] command failed: %s/%s err=%v", req.Module, req.Cmd, err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Infof("[debug] command executed: %s/%s params=%v result=%+v", req.Module, req.Cmd, req.Params, ret)
	writeJSON(w, http.StatusOK, ret)
}

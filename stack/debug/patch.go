package debug

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/ddd"
)

// patchRequest 直接修改聚合字段的请求体。
type patchRequest struct {
	Module string         `json:"module" example:"player"`
	ID     int64          `json:"id"     example:"12345"`
	Fields map[string]any `json:"fields"`
}

// @Summary      直接修改聚合字段
// @Description  用 unsafe 绕过导出限制直接写内存。不经过构造校验，不维护不变量。仅用于调试。
// @Description  写操作不走 Actor 串行化，玩家在线时可能产生竞态。
// @Tags         patch
// @Accept       json
// @Produce      json
// @Param        request  body  patchRequest  true  "patch 请求"
// @Success      200      {object}  map[string]any  "修改后的快照"
// @Failure      400      {object}  map[string]string  "字段不存在或类型错误"
// @Failure      404      {object}  map[string]string  "模块或聚合不存在"
// @Router       /debug/patch [post]
func handlePatch(w http.ResponseWriter, r *http.Request) {
	var req patchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	if req.Module == "" || req.ID == 0 || len(req.Fields) == 0 {
		writeError(w, http.StatusBadRequest, "module, id, and fields are required")
		return
	}

	m, ok := get(req.Module)
	if !ok {
		writeError(w, http.StatusNotFound, "module not found: "+req.Module)
		return
	}

	agg, err := m.Load(context.Background(), req.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, req.Module+" "+strconv.FormatInt(req.ID, 10)+" not found: "+err.Error())
		return
	}

	if err := ddd.ApplyPatch(agg, req.Fields); err != nil {
		writeError(w, http.StatusBadRequest, "patch failed: "+err.Error())
		return
	}

	if err := m.Save(context.Background(), agg); err != nil {
		writeError(w, http.StatusInternalServerError, "save failed: "+err.Error())
		return
	}

	patched := make([]string, 0, len(req.Fields))
	for k := range req.Fields {
		patched = append(patched, k)
	}

	log.Infof("[debug] patched %s/%d fields=%v", req.Module, req.ID, patched)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"patched":  patched,
		"snapshot": ddd.Snapshot(agg),
	})
}

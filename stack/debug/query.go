package debug

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/ddd"
)

// queryRequest 查询聚合的请求体。
type queryRequest struct {
	Module string `json:"module" example:"player"`
	ID     int64  `json:"id"     example:"12345"`
}

// @Summary      查询聚合完整快照
// @Description  返回聚合所有非导出字段的当前值。值对象自动展开为原始类型（Gold(999) → 999）。
// @Tags         query
// @Accept       json
// @Produce      json
// @Param        request  body  queryRequest  true  "查询请求"
// @Success      200      {object}  map[string]any  "聚合快照"
// @Failure      400      {object}  map[string]string  "参数错误"
// @Failure      404      {object}  map[string]string  "模块或聚合不存在"
// @Router       /debug/query [post]
func handleQuery(w http.ResponseWriter, r *http.Request) {
	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	if req.Module == "" || req.ID == 0 {
		writeError(w, http.StatusBadRequest, "module and id are required")
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

	snap := ddd.Snapshot(agg)
	log.Infof("[debug] queried %s/%d", req.Module, req.ID)
	writeJSON(w, http.StatusOK, snap)
}

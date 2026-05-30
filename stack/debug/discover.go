package debug

import (
	"net/http"
)

// @Summary      列出所有已注册模块
// @Description  返回所有已通过 debug.Register 注册的模块名称列表。
// @Tags         discover
// @Produce      json
// @Success      200  {array}   string  "模块名称列表"
// @Router       /debug/modules [get]
func handleDiscover(w http.ResponseWriter, r *http.Request) {
	ns := names()
	writeJSON(w, http.StatusOK, ns)
}

// @Summary      列出模块能力
// @Description  返回指定模块的查询能力和可用命令列表。
// @Tags         discover
// @Produce      json
// @Param        name  path  string  true  "模块名称（如 player）"
// @Success      200   {object}  map[string]any
// @Failure      404   {object}  map[string]string  "模块不存在"
// @Router       /debug/module/{name} [get]
func handleModuleInfo(w http.ResponseWriter, r *http.Request, name string) {
	m, ok := get(name)
	if !ok {
		writeError(w, http.StatusNotFound, "module not found: "+name)
		return
	}

	info := map[string]any{
		"queries": []string{"get"},
	}
	if m.CmdBus != nil {
		info["commands"] = m.CmdBus.CommandNames()
	}
	writeJSON(w, http.StatusOK, info)
}

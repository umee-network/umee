package swagger

import (
	"embed"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ignite/cli/ignite/pkg/openapiconsole"
)

//go:embed swagger.yaml
var Docs embed.FS

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(rtr *mux.Router) {
	rtr.Handle("/swagger.yaml", http.FileServer(http.FS(Docs)))
	rtr.HandleFunc("/swagger/", openapiconsole.Handler("umee", "/swagger.yaml"))
}

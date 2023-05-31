package swagger

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"

	// unnamed import of statik for swagger UI support
	_ "github.com/umee-network/umee/v5/swagger/statik"
)

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

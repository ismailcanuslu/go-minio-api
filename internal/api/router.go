package api

import "net/http"

func NewRouter(controller *Controller) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", controller.Health)
	mux.HandleFunc("GET /objects", controller.ListObjects)
	mux.HandleFunc("PUT /objects/", controller.PutObject)
	mux.HandleFunc("GET /objects/", controller.GetObject)
	mux.HandleFunc("DELETE /objects/", controller.DeleteObject)
	mux.HandleFunc("GET /objects-meta/", controller.GetObjectMeta)
	mux.HandleFunc("POST /presign/", controller.PresignGetObject)

	return loggingMiddleware(mux)
}

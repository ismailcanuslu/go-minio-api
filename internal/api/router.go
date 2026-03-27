package api

import "net/http"

func NewRouter(controller *Controller) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", controller.Health)

	// Bucket CRUD
	mux.HandleFunc("GET /buckets", controller.ListBuckets)
	mux.HandleFunc("POST /buckets/{name}", controller.CreateBucket)
	mux.HandleFunc("DELETE /buckets/{name}", controller.DeleteBucket)

	// Object operations — all scoped to a bucket
	mux.HandleFunc("GET /buckets/{name}/objects", controller.ListObjects)
	mux.HandleFunc("PUT /buckets/{name}/objects/", controller.PutObject)
	mux.HandleFunc("GET /buckets/{name}/objects/", controller.GetObject)
	mux.HandleFunc("DELETE /buckets/{name}/objects/", controller.DeleteObject)
	mux.HandleFunc("GET /buckets/{name}/meta/", controller.GetObjectMeta)
	mux.HandleFunc("POST /buckets/{name}/presign/", controller.PresignGetObject)

	return loggingMiddleware(mux)
}

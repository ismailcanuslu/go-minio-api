package api

import (
	"context"
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"

	"go-file-microservice/internal/storage"
)

type Controller struct {
	storage *storage.MinIOService
}

func NewController(storageService *storage.MinIOService) *Controller {
	return &Controller{storage: storageService}
}

func (c *Controller) Health(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (c *Controller) ListObjects(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	recursive := r.URL.Query().Get("recursive") != "false"

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	objects, err := c.storage.ListObjects(ctx, prefix, recursive)
	if err != nil {
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"bucket":    c.storage.Bucket(),
		"prefix":    prefix,
		"recursive": recursive,
		"count":     len(objects),
		"objects":   objects,
	})
}

func (c *Controller) PutObject(w http.ResponseWriter, r *http.Request) {
	key := objectKey(r.URL.Path, "/objects/")
	if key == "" {
		respondError(w, http.StatusBadRequest, "object key is required")
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(extFromKey(key))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	info, err := c.storage.PutObject(ctx, key, contentType, r.Body, -1)
	if err != nil {
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"bucket":      c.storage.Bucket(),
		"key":         key,
		"etag":        info.ETag,
		"size":        info.Size,
		"contentType": contentType,
	})
}

func (c *Controller) GetObject(w http.ResponseWriter, r *http.Request) {
	key := objectKey(r.URL.Path, "/objects/")
	if key == "" {
		respondError(w, http.StatusBadRequest, "object key is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	obj, stat, err := c.storage.GetObject(ctx, key)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	defer obj.Close()

	if stat.ContentType != "" {
		w.Header().Set("Content-Type", stat.ContentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size, 10))
	w.Header().Set("ETag", stat.ETag)
	w.Header().Set("Last-Modified", stat.LastModified.UTC().Format(http.TimeFormat))

	if _, err = io.Copy(w, obj); err != nil {
		respondError(w, http.StatusInternalServerError, "stream object failed")
	}
}

func (c *Controller) DeleteObject(w http.ResponseWriter, r *http.Request) {
	key := objectKey(r.URL.Path, "/objects/")
	if key == "" {
		respondError(w, http.StatusBadRequest, "object key is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	if err := c.storage.DeleteObject(ctx, key); err != nil {
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"bucket": c.storage.Bucket(),
		"key":    key,
		"status": "deleted",
	})
}

func (c *Controller) GetObjectMeta(w http.ResponseWriter, r *http.Request) {
	key := objectKey(r.URL.Path, "/objects-meta/")
	if key == "" {
		respondError(w, http.StatusBadRequest, "object key is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	objectInfo, err := c.storage.StatObject(ctx, key)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, objectInfo)
}

func (c *Controller) PresignGetObject(w http.ResponseWriter, r *http.Request) {
	key := objectKey(r.URL.Path, "/presign/")
	if key == "" {
		respondError(w, http.StatusBadRequest, "object key is required")
		return
	}

	expiry := 15 * time.Minute
	if value := r.URL.Query().Get("expiryMinutes"); value != "" {
		minutes, err := strconv.Atoi(value)
		if err != nil || minutes <= 0 || minutes > 60*24*7 {
			respondError(w, http.StatusBadRequest, "expiryMinutes must be between 1 and 10080")
			return
		}
		expiry = time.Duration(minutes) * time.Minute
	}

	url, err := c.storage.PresignGetObject(r.Context(), key, expiry)
	if err != nil {
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"bucket":         c.storage.Bucket(),
		"key":            key,
		"expiresIn":      expiry.String(),
		"presignedGet":   url,
		"recommendedUse": "temporary public download",
	})
}

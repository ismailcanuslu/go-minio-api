## Go File Microservice + MinIO

Bu proje, MinIO üzerinde dosya işlemleri yapan bir Go mikroservisidir.

### Özellikler

- İlgili dizin/prefix altındaki dosyaları listeler
- Dosya oluşturur / günceller (upload)
- Dosya içeriğini servis eder (download/stream)
- Dosya metadata bilgisini döner
- Dosya siler
- Geçici erişim için presigned URL üretir

## Çalıştırma

```bash
docker compose up --build
```

- API: `http://localhost:8080`
- MinIO API: `http://localhost:9000`
- MinIO Console: `http://localhost:9001` (kullanıcı/şifre: `minioadmin`)

## API Rotaları

### Controller Noktaları

| Method | Route | Controller Metodu | Katman |
|---|---|---|---|
| GET | `/health` | `Controller.Health` | `internal/api/controller.go` |
| GET | `/objects` | `Controller.ListObjects` | `internal/api/controller.go` |
| PUT | `/objects/{key}` | `Controller.PutObject` | `internal/api/controller.go` |
| GET | `/objects/{key}` | `Controller.GetObject` | `internal/api/controller.go` |
| DELETE | `/objects/{key}` | `Controller.DeleteObject` | `internal/api/controller.go` |
| GET | `/objects-meta/{key}` | `Controller.GetObjectMeta` | `internal/api/controller.go` |
| POST | `/presign/{key}` | `Controller.PresignGetObject` | `internal/api/controller.go` |

Router tanımları `internal/api/router.go` dosyasındadır.

### Health

```bash
curl http://localhost:8080/health
```

### Dizin/PREFIX Listeleme

```bash
curl "http://localhost:8080/objects?prefix=docs/&recursive=true"
```

### Dosya Oluşturma (Upload)

```bash
curl -X PUT \
  -H "Content-Type: text/plain" \
  --data-binary "Merhaba MinIO" \
  http://localhost:8080/objects/docs/hello.txt
```

### Dosya İçeriği Servis Etme

```bash
curl http://localhost:8080/objects/docs/hello.txt
```

### Dosya Metadata

```bash
curl http://localhost:8080/objects-meta/docs/hello.txt
```

### Dosya Silme

```bash
curl -X DELETE http://localhost:8080/objects/docs/hello.txt
```

### Presigned Download URL

```bash
curl -X POST "http://localhost:8080/presign/docs/hello.txt?expiryMinutes=30"
```

## Ortam Değişkenleri

- `SERVER_PORT` (varsayılan: `8080`)
- `MINIO_ENDPOINT` (varsayılan: `minio:9000`)
- `MINIO_ACCESS_KEY` (varsayılan: `minioadmin`)
- `MINIO_SECRET_KEY` (varsayılan: `minioadmin`)
- `MINIO_USE_SSL` (varsayılan: `false`)
- `MINIO_BUCKET` (varsayılan: `app-files`)

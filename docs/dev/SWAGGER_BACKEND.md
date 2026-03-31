# Backend Swagger（Gin + swaggo）

## 生成

在 PowerShell 中执行：

```powershell
cd d:\Projects\irms\backend
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init `
  -g cmd/irms/main.go `
  -o .\docs\swagger `
  --parseInternal `
  --dir .
```

生成物：

- `backend/docs/swagger/swagger.json`
- `backend/docs/swagger/swagger.yaml`
- `backend/docs/swagger/docs.go`

## 访问

启动后端（示例）：

```powershell
cd d:\Projects\irms\backend
go run .\cmd\irms
```

浏览器访问：

- `http://127.0.0.1:8080/api/swagger/index.html`

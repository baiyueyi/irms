# Rule 02｜工作区结构

- `D:\Projects\irms` 是工作区根目录，不是前端或后端单独根目录。
- `.trae/`：Trae 规则、spec 任务板。
- `docs/`：正式文档与验收证据。
- `frontend/`：前端项目根目录。
- `backend/`：后端项目根目录。
- 后端 Go 相关内容必须放在 `backend/` 内，例如：
  - `backend/go.mod`
  - `backend/go.sum`
  - `backend/cmd/`
  - `backend/internal/`
- 不允许把工作区根目录和后端项目根目录混用。

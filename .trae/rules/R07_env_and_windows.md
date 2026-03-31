# Rule 07｜环境变量与 Windows

- 当前开发环境固定为 Windows。
- 所有命令、脚本、路径写法优先兼容 PowerShell。
- 不使用只适配 bash 的命令写法。
- 不假设 Linux 目录结构。
- 现有环境变量直接使用，不引入 dotenv。
- 不覆盖以下变量：
  - `CREDENTIAL_ENCRYPTION_KEY`
  - `JWT_SECRET`
  - `DB_HOST`
  - `DB_NAME`
  - `DB_PASSWORD`
  - `DB_USER`
  - `DB_PORT`
- 生成 PowerShell 命令时，避免单引号嵌套问题。

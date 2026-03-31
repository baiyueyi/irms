# Rule 11｜输出与验收

- 中大型任务每次回复固定输出：
  - A) Current Phase
  - B) Status
  - C) What Changed
  - D) Next Actions
- 小型 bugfix 可简化文字，但必须交代改动、验证、剩余风险。
- 输出简洁，不写长篇空话。
- 完成标准不是“页面能打开”或“build 通过”。
- 完成必须同时满足：
  - 代码已落地
  - 构建通过
  - 关键链路已手验
  - 文档已落盘
  - 缺口与限制已记录
- 验证优先使用 PowerShell 步骤。
- 验收证据写入 `docs/qa/**` 与 `docs/releases/**`。

# Rule 05｜任务板与正式文档

- `.trae/specs/**` 只放任务板：
  - `spec.md`
  - `tasks.md`
  - `checklist.md`
- `docs/**` 只放正式文档与证据。
- 正式文档目录：
  - `docs/product/**`
  - `docs/arch/**`
  - `docs/dev/**`
  - `docs/qa/**`
  - `docs/releases/**`
- 不要一次性铺满所有文档，只生成当前阶段真正需要的文档。
- 不要生成大量给 AI 自己看的冗长中间文档。
- 规则、契约、迁移说明优先收口到 `docs/dev/**` 或 `docs/arch/**`，不要散落在临时输出里。

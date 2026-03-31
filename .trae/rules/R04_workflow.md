# Rule 04｜工作流顺序

- 固定顺序：
  1. Clarify
  2. Planning
  3. Audit
  4. Gap List
  5. Minimal Plan
  6. Implement
  7. Verify
  8. Freeze
- 需求不清晰时先提问，不直接写代码。
- 重要任务先建立 `.trae/specs/<topic>/` 任务板。
- 审计未完成前，不直接大改代码。
- 先做最小可用实现，再做阻塞修复，再做收口。
- 小型 bugfix 可简化任务板，但不得跳过审计与验证。
- 禁止把中等需求自动升级成平台级重构。

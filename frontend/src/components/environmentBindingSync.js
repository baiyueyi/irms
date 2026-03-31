export function diffEnvironmentIds(previousIds, nextIds) {
  const prev = Array.from(new Set((previousIds || []).map((x) => Number(x))))
  const next = Array.from(new Set((nextIds || []).map((x) => Number(x))))
  const prevSet = new Set(prev)
  const nextSet = new Set(next)
  const added = next.filter((id) => !prevSet.has(id))
  const removed = prev.filter((id) => !nextSet.has(id))
  return { added, removed }
}

// Stars renders a rating out of 5. value may be fractional (e.g. the average).
export default function Stars({ value }) {
  const rounded = Math.round(Number(value) || 0)
  return (
    <span className="whitespace-nowrap text-base tracking-wider text-amber-500" aria-label={`${rounded} out of 5 stars`}>
      {'★'.repeat(rounded)}
      <span className="text-slate-300 dark:text-slate-700">{'★'.repeat(5 - rounded)}</span>
    </span>
  )
}

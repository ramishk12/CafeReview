// Stars renders a rating out of 5. value may be fractional (e.g. the average).
export default function Stars({ value }) {
  const rounded = Math.round(Number(value) || 0)
  return (
    <span className="stars" aria-label={`${rounded} out of 5 stars`}>
      {'★'.repeat(rounded)}
      <span className="stars-empty">{'★'.repeat(5 - rounded)}</span>
    </span>
  )
}

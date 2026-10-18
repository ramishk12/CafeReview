// Shared loading placeholders. The route-level loading.jsx files compose these,
// so every page waits with the same skeleton style.

// A single pulsing block. Size and shape it with className.
export function Skeleton({ className = '' }) {
  return <div aria-hidden="true" className={`animate-pulse rounded-2xl bg-slate-200/80 dark:bg-slate-800/80 ${className}`} />
}

// A page with a heading and a grid of cards (the cafe list, My lists).
export function PageSkeleton({ cards = 6 }) {
  return (
    <div aria-busy="true" aria-label="Loading" className="space-y-8 pt-4">
      <div className="space-y-3">
        <Skeleton className="h-3 w-20 rounded-full" />
        <Skeleton className="h-10 w-3/4" />
        <Skeleton className="h-4 w-1/2" />
      </div>
      <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
        {Array.from({ length: cards }, (_, i) => (
          <Skeleton key={i} className="h-56" />
        ))}
      </div>
    </div>
  )
}

// A detail page: a header block followed by content rows (cafe, profile, list).
export function DetailSkeleton({ rows = 3 }) {
  return (
    <div aria-busy="true" aria-label="Loading" className="space-y-6 pt-4">
      <Skeleton className="h-4 w-24 rounded-full" />
      <Skeleton className="h-48" />
      <div className="space-y-4">
        {Array.from({ length: rows }, (_, i) => (
          <Skeleton key={i} className="h-28" />
        ))}
      </div>
    </div>
  )
}

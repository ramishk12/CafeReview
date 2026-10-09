// Skeleton for the cafe list while its data loads. Scoped to "/" so cafe detail pages return a real 404 when missing.
export default function Loading() {
  return (
    <div aria-busy="true" aria-label="Loading" className="animate-pulse space-y-8 pt-4">
      <div className="space-y-3">
        <div className="h-3 w-20 rounded bg-slate-200 dark:bg-slate-800" />
        <div className="h-10 w-3/4 rounded bg-slate-200 dark:bg-slate-800" />
        <div className="h-4 w-1/2 rounded bg-slate-200 dark:bg-slate-800" />
      </div>
      <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
        {Array.from({ length: 6 }, (_, i) => (
          <div key={i} className="card h-56 bg-slate-100 dark:bg-slate-900" />
        ))}
      </div>
    </div>
  )
}

// Shared empty state: an optional icon, a short title, one line of guidance,
// and an optional action (a button or link) that says what to do next.
//
// titleAs lets a page keep a sensible heading level, for example an h1 on the
// error page.
export default function EmptyState({ icon: Icon, title, description, action, titleAs: Title = 'h2', className = '' }) {
  return (
    <div className={`card flex flex-col items-center gap-3 px-6 py-14 text-center ${className}`}>
      {Icon && (
        <span
          aria-hidden="true"
          className="mb-1 flex h-11 w-11 items-center justify-center rounded-2xl bg-slate-100 text-slate-500 dark:bg-white/10 dark:text-slate-300"
        >
          <Icon size={20} />
        </span>
      )}
      <Title className="text-lg font-semibold tracking-tight">{title}</Title>
      {description && <p className="max-w-md text-sm text-slate-500 dark:text-slate-400">{description}</p>}
      {action && <div className="mt-2">{action}</div>}
    </div>
  )
}

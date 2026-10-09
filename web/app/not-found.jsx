import Link from 'next/link'

export const metadata = { title: 'Not found' }

export default function NotFound() {
  return (
    <div className="card mx-auto mt-12 max-w-md p-8 text-center">
      <p className="eyebrow">404</p>
      <h1 className="mt-2 text-2xl font-bold tracking-tight">We couldn&apos;t find that page</h1>
      <p className="mt-2 text-slate-500 dark:text-slate-400">It may have been moved or never existed.</p>
      <Link href="/" className="btn mt-6">
        Back to cafes
      </Link>
    </div>
  )
}

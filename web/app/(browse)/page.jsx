import { listCafes } from '@/lib/serverApi'
import CafeBrowser from '@/components/CafeBrowser'

export const metadata = {
  title: 'Cafes',
  description: 'Browse cafes with honest reviews and photos.',
}

export default async function HomePage() {
  const cafes = await listCafes()
  const reviewTotal = cafes.reduce((sum, c) => sum + c.review_count, 0)

  return (
    <section>
      <div className="mb-10 pt-4">
        <p className="eyebrow">Discover</p>
        <h1 className="mt-2 text-4xl font-extrabold tracking-tight sm:text-5xl">
          Find your next favorite{' '}
          <span className="bg-linear-to-r from-indigo-600 to-fuchsia-500 bg-clip-text text-transparent">cafe</span>
        </h1>
        <p className="mt-3 max-w-xl text-lg text-slate-500 dark:text-slate-400">
          Honest reviews and photos from people who actually drink the coffee.
        </p>
        <p className="mt-4 text-sm text-slate-500 dark:text-slate-400">
          {cafes.length} {cafes.length === 1 ? 'cafe' : 'cafes'} · {reviewTotal} {reviewTotal === 1 ? 'review' : 'reviews'}
        </p>
      </div>

      <CafeBrowser cafes={cafes} />
    </section>
  )
}

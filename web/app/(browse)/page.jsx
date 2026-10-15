import { listCafes, listTags } from '@/lib/serverApi'
import CafeBrowser from '@/components/CafeBrowser'

export const metadata = {
  title: 'Cafes',
  description: 'Browse cafes with honest reviews and photos.',
}

const PAGE_SIZE = 24
const SORTS = ['name', 'rating', 'reviews', 'newest']

export default async function HomePage({ searchParams }) {
  const sp = await searchParams
  const page = Math.max(1, Number(sp.page) || 1)
  const q = typeof sp.q === 'string' ? sp.q : ''
  const sort = SORTS.includes(sp.sort) ? sp.sort : 'rating'
  const tags = typeof sp.tags === 'string' ? sp.tags : ''

  const [{ cafes, total }, allTags] = await Promise.all([
    listCafes({ q, sort, tags, page, pageSize: PAGE_SIZE }),
    listTags(),
  ])

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
          {total} {total === 1 ? 'cafe' : 'cafes'}
        </p>
      </div>

      <CafeBrowser cafes={cafes} tags={allTags} total={total} page={page} pageSize={PAGE_SIZE} />
    </section>
  )
}

import { MessageSquare } from 'lucide-react'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import AddToListButton from '@/components/AddToListButton'
import EmptyState from '@/components/EmptyState'
import Stars from '@/components/Stars'
import { getUser, listUserReviews } from '@/lib/serverApi'
import { avatarColor, formatDate, initials } from '@/lib/cafeTheme'

const PAGE_SIZE = 20

export async function generateMetadata({ params }) {
  const { id } = await params
  const user = await getUser(id).catch(() => null)
  return { title: user ? user.display_name : 'Profile' }
}

export default async function ProfilePage({ params, searchParams }) {
  const { id } = await params
  if (!/^\d+$/.test(id)) notFound()

  const user = await getUser(id)
  if (!user) notFound()

  const sp = await searchParams
  const page = Math.max(1, Number(sp.page) || 1)
  const { reviews, total } = await listUserReviews(id, { page, pageSize: PAGE_SIZE })
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return (
    <section className="mx-auto max-w-3xl">
      <header className="card mb-8 flex items-center gap-4 p-6">
        <span
          className={`flex h-14 w-14 items-center justify-center rounded-full text-lg font-semibold text-white shadow-inner ${avatarColor(
            user.display_name,
          )}`}
          aria-hidden="true"
        >
          {initials(user.display_name)}
        </span>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{user.display_name}</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400">
            Joined {formatDate(user.created_at)} · {user.review_count} {user.review_count === 1 ? 'review' : 'reviews'}
          </p>
        </div>
        <div className="ml-auto">
          <AddToListButton type="user" id={user.id} label="Save to list" />
        </div>
      </header>

      {reviews.length === 0 ? (
        <EmptyState
          icon={MessageSquare}
          title="No reviews yet"
          description={`${user.display_name} hasn't reviewed any cafes yet.`}
          action={
            <Link href="/" className="btn">
              Browse cafes
            </Link>
          }
        />
      ) : (
        <ul className="grid gap-4">
          {reviews.map((review) => (
            <li key={review.id} className="card p-6">
              <div className="flex items-center justify-between gap-3">
                <Link href={`/cafes/${review.cafe_id}`} className="font-semibold no-underline">
                  {review.cafe_name}
                </Link>
                <Stars value={review.rating} size={15} />
              </div>
              <time dateTime={review.visited_on || review.created_at} className="text-xs text-slate-400">
                {review.visited_on ? `Visited ${formatDate(review.visited_on)}` : formatDate(review.created_at)}
              </time>
              <p className="mt-3 whitespace-pre-wrap leading-relaxed text-slate-700 dark:text-slate-300">{review.body}</p>
              {review.tags?.length > 0 && (
                <div className="mt-3 flex flex-wrap gap-2">
                  {review.tags.map((tag) => (
                    <span key={tag.id} className="rounded-full bg-slate-100 px-2.5 py-0.5 text-xs font-medium text-slate-600 dark:bg-white/10 dark:text-slate-300">
                      {tag.name}
                    </span>
                  ))}
                </div>
              )}
              {review.images?.length > 0 && (
                <ul className="mt-3 flex flex-wrap gap-2">
                  {review.images.map((img) => (
                    <li key={img.id}>
                      <img src={img.url} alt="" loading="lazy" className="h-20 w-20 rounded-xl object-cover" />
                    </li>
                  ))}
                </ul>
              )}
            </li>
          ))}
        </ul>
      )}

      {totalPages > 1 && (
        <nav className="mt-8 flex items-center justify-center gap-3" aria-label="Pagination">
          {page > 1 && (
            <Link href={page - 1 === 1 ? `/users/${id}` : `/users/${id}?page=${page - 1}`} className="btn btn-ghost">
              Previous
            </Link>
          )}
          <span className="px-2 text-sm text-slate-500 dark:text-slate-400">
            Page {page} of {totalPages}
          </span>
          {page < totalPages && (
            <Link href={`/users/${id}?page=${page + 1}`} className="btn btn-ghost">
              Next
            </Link>
          )}
        </nav>
      )}
    </section>
  )
}

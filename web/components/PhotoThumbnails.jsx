'use client'

import { useState } from 'react'
import PhotoLightbox from './PhotoLightbox'

// Small photo strip for pages that only show thumbnails (the profile page).
// Clicking a thumbnail opens the shared lightbox.
export default function PhotoThumbnails({ images, authorName }) {
  const [index, setIndex] = useState(null)
  const photos = (images || []).map((img) => ({ id: img.id, url: img.url, alt: `Photo by ${authorName}` }))

  if (photos.length === 0) return null

  return (
    <>
      <ul className="mt-3 flex flex-wrap gap-2">
        {photos.map((photo, i) => (
          <li key={photo.id}>
            <button
              type="button"
              onClick={() => setIndex(i)}
              className="block cursor-zoom-in overflow-hidden rounded-xl"
              aria-label={`Open ${photo.alt.toLowerCase()}`}
            >
              <img src={photo.url} alt="" loading="lazy" className="h-20 w-20 object-cover" />
            </button>
          </li>
        ))}
      </ul>
      <PhotoLightbox images={photos} index={index} onIndexChange={setIndex} onClose={() => setIndex(null)} />
    </>
  )
}

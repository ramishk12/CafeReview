import '@fontsource-variable/inter'
import './globals.css'
import Navbar from '@/components/Navbar'
import Providers from './providers'

export const metadata = {
  title: { default: 'CafeReview', template: '%s · CafeReview' },
  description: 'Honest reviews and photos for the cafes you love.',
}

export const viewport = {
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#f8fafc' },
    { media: '(prefers-color-scheme: dark)', color: '#020617' },
  ],
}

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>
        <Providers>
          <Navbar />
          <main className="mx-auto max-w-5xl px-5 pt-8 pb-16">{children}</main>
        </Providers>
      </body>
    </html>
  )
}

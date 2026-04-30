// A11Y ISSUE: html tag missing lang attribute (WCAG_3_1_1)
export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <header>
          <nav aria-label="Main Navigation">
            <a href="/">Home</a>
            <a href="/about">About</a>
            <a href="/dashboard">Dashboard</a>
          </nav>
        </header>
        <main>{children}</main>
        {/* A11Y ISSUE: low contrast footer text */}
        <footer>
          <p>© 2024 Example Next.js App</p>
        </footer>
      </body>
    </html>
  );
}

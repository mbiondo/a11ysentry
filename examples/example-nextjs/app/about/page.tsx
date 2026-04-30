export default function AboutPage() {
  return (
    <>
      <h1>About</h1>
      <p>Built with Next.js App Router.</p>

      {/* A11Y ISSUE: button missing accessible name (WCAG_4_1_2) */}
      <button onClick={() => {}}>
        <svg viewBox="0 0 24 24" width="24" height="24" aria-hidden="true">
          <path d="M19 11H7.83l4.88-4.88c.39-.39.39-1.03 0-1.42-.39-.39-1.02-.39-1.41 0l-6.59 6.59c-.39.39-.39 1.02 0 1.41l6.59 6.59c.39.39 1.02.39 1.41 0 .39-.39.39-1.02 0-1.41L7.83 13H19c.55 0 1-.45 1-1s-.45-1-1-1z"/>
        </svg>
      </button>

      {/* A11Y ISSUE: heading jump h1 -> h3 (WCAG_1_3_1) */}
      <h3>Our Story</h3>
      <p>We started building accessible apps in 2020.</p>
    </>
  );
}

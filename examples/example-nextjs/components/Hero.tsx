export default function Hero() {
  return (
    <section>
      {/* A11Y ISSUE: image missing alt (WCAG_1_1_1) */}
      <img src="/hero.png" alt="Hero image" />
      <p>Build fast, scalable web applications.</p>
      <label htmlFor="newsletter-email">Email</label>
      <input id="newsletter-email" type="email" placeholder="Enter your email" />
      <button>Get Started</button>
    </section>
  );
}

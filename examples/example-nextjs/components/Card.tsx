interface CardProps {
  title: string;
  value: string;
}

export default function Card({ title, value }: CardProps) {
  return (
    <article className="card">
      <h3 className="card-title">{title}</h3>
      <p className="card-value">{value}</p>
  /* Fixed a11y: ensure sufficient color contrast for the button text */
  <button style={{ color: '#111111', backgroundColor: '#ffffff' }}>View details</button>
    </article>
  );
}

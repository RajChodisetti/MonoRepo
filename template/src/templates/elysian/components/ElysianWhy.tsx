import type { ExperienceCard } from "@/data/types/restaurant";

const ICONS = [
  <svg key="0" viewBox="0 0 24 24"><path d="M4 20V10M4 10 12 4l8 6M4 10h16M8 20v-5a4 4 0 0 1 8 0v5" stroke="currentColor" strokeWidth="1.4" fill="none" /></svg>,
  <svg key="1" viewBox="0 0 24 24"><path d="M12 2v6M12 16v6M4.9 4.9l4.2 4.2M14.9 14.9l4.2 4.2M2 12h6M16 12h6M4.9 19.1l4.2-4.2M14.9 9.1l4.2-4.2" stroke="currentColor" strokeWidth="1.4" /></svg>,
  <svg key="2" viewBox="0 0 24 24"><path d="M12 3C7 8 5 11 5 14a7 7 0 0 0 14 0c0-3-2-6-7-11Z" stroke="currentColor" strokeWidth="1.4" fill="none" /></svg>,
  <svg key="3" viewBox="0 0 24 24"><path d="M12 7v5l3 3M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Z" stroke="currentColor" strokeWidth="1.4" fill="none" /></svg>,
  <svg key="4" viewBox="0 0 24 24"><path d="M4 10h16M6 10v10h12V10M9 21v-6h6v6" stroke="currentColor" strokeWidth="1.4" fill="none" /></svg>,
  <svg key="5" viewBox="0 0 24 24"><path d="M5 3h14l-1 8a6 6 0 0 1-12 0L5 3ZM12 17v4M8 21h8" stroke="currentColor" strokeWidth="1.4" fill="none" /></svg>,
];

export default function ElysianWhy({ cards }: { cards: ExperienceCard[] }) {
  if (!cards.length) return null;

  return (
    <section className="why section" id="why">
      <div className="container">
        <div className="section-head reveal fade-up">
          <p className="eyebrow">Why Visit</p>
          <h2 className="section-title">
            The Details That <span className="gold-text">Define Us</span>
          </h2>
        </div>
        <div className="why-grid">
          {cards.map((card, i) => (
            <div key={card.id} className="why-card reveal fade-up glass">
              <span className="why-icon">{ICONS[i % ICONS.length]}</span>
              <h3>{card.title}</h3>
              <p>{card.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

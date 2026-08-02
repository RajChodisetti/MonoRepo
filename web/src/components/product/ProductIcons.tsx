import type { JSX } from "react";
import type { IconId } from "@/content/products/types";

const iconClass = "h-5 w-5";

function BadgeIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <path
        d="M12 2.8 14.2 5l2.8.4-2 2.2.5 2.8L12 9.2 8.5 10.4l.5-2.8-2-2.2L9.8 5 12 2.8Z"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinejoin="round"
      />
      <path
        d="M9.2 14.2 11 16l3.8-4"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M7.5 13.5v6.2L12 17.8l4.5 1.9v-6.2"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function PercentIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <circle cx="8" cy="8" r="2.2" fill="none" stroke="currentColor" strokeWidth="1.5" />
      <circle cx="16" cy="16" r="2.2" fill="none" stroke="currentColor" strokeWidth="1.5" />
      <path d="M17.5 6.5 6.5 17.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

function GaugeIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <path
        d="M5.2 16.2a7.8 7.8 0 1 1 13.6 0"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
      <path d="M12 14.5 15.2 9.8" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
      <circle cx="12" cy="14.8" r="1.3" fill="currentColor" />
    </svg>
  );
}

function ChartIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <path
        d="M4 19h16M7 16V10M12 16V7M17 16v-4"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function UsersIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <circle cx="9" cy="8" r="2.5" fill="none" stroke="currentColor" strokeWidth="1.5" />
      <circle cx="16" cy="9" r="2" fill="none" stroke="currentColor" strokeWidth="1.5" />
      <path
        d="M4.5 18c.6-2.4 2.4-3.5 4.5-3.5s3.9 1.1 4.5 3.5"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
      <path
        d="M14.2 14.2c1.5.2 2.9 1.1 3.5 2.8"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
    </svg>
  );
}

function PencilIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <path
        d="m14.2 5.2 4.6 4.6M5 19l1.2-4.4L16.6 4.2a1.6 1.6 0 0 1 2.3 0l.9.9a1.6 1.6 0 0 1 0 2.3L9.4 17.8 5 19Z"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function DiamondIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <path
        d="M12 3.5 4.8 10.2 12 20.5l7.2-10.3L12 3.5Z"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinejoin="round"
      />
      <path
        d="M4.8 10.2h14.4M8.2 10.2 12 3.5l3.8 6.7M8.2 10.2 12 20.5l3.8-10.3"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function PersonIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <circle cx="12" cy="12" r="8.2" fill="none" stroke="currentColor" strokeWidth="1.5" />
      <circle cx="12" cy="10" r="2.4" fill="none" stroke="currentColor" strokeWidth="1.5" />
      <path
        d="M7.8 17.2c.9-1.8 2.4-2.7 4.2-2.7s3.3.9 4.2 2.7"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
    </svg>
  );
}

function GearIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" strokeWidth="1.5" />
      <path
        d="M12 2.8v2.2M12 19v2.2M2.8 12h2.2M19 12h2.2M5.1 5.1l1.6 1.6M17.3 17.3l1.6 1.6M5.1 18.9l1.6-1.6M17.3 6.7l1.6-1.6"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
      <path
        d="M9.4 12.2 11.2 14l3.6-3.8"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function BoltIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <path
        d="M13.2 2.8 6.5 13.2h4.8l-.8 8 6.8-10.6h-4.7l.6-7.8Z"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function WalletIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <path
        d="M4 7.5h14.5A1.5 1.5 0 0 1 20 9v9.5a1.5 1.5 0 0 1-1.5 1.5H5.5A1.5 1.5 0 0 1 4 18.5V7.5Z"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinejoin="round"
      />
      <path
        d="M4 7.5V6.2A1.7 1.7 0 0 1 5.7 4.5H16"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
      <circle cx="16.2" cy="14" r="1.2" fill="currentColor" />
    </svg>
  );
}

function CarIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <path
        d="M4 15.5h16v3.2H4v-3.2ZM5.5 15.5l1.4-5.2A1.5 1.5 0 0 1 8.3 9h7.4a1.5 1.5 0 0 1 1.4 1.3l1.4 5.2"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinejoin="round"
      />
      <circle cx="7.5" cy="18.2" r="1.2" fill="currentColor" />
      <circle cx="16.5" cy="18.2" r="1.2" fill="currentColor" />
    </svg>
  );
}

function PhoneIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <path
        d="M8.2 4.5h7.6A1.7 1.7 0 0 1 17.5 6.2v11.6a1.7 1.7 0 0 1-1.7 1.7H8.2a1.7 1.7 0 0 1-1.7-1.7V6.2A1.7 1.7 0 0 1 8.2 4.5Z"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
      />
      <path d="M10.5 17.5h3" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

function TrophyIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <path
        d="M8 4.5h8v4.2a4 4 0 0 1-8 0V4.5ZM6 5.5H4.5A2 2 0 0 0 4.5 9c1.2 0 1.5-.8 1.5-1.6V5.5ZM18 5.5h1.5A2 2 0 0 1 19.5 9c-1.2 0-1.5-.8-1.5-1.6V5.5ZM10 16.5h4M12 12.7v3.8M9 19.5h6"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function CardIcon() {
  return (
    <svg viewBox="0 0 24 24" className={iconClass} aria-hidden="true">
      <rect
        x="3.5"
        y="5.5"
        width="17"
        height="13"
        rx="2"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
      />
      <path d="M3.5 9.5h17M8 14h4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

const ICON_MAP: Record<IconId, () => JSX.Element> = {
  badge: BadgeIcon,
  percent: PercentIcon,
  gauge: GaugeIcon,
  chart: ChartIcon,
  users: UsersIcon,
  pencil: PencilIcon,
  diamond: DiamondIcon,
  person: PersonIcon,
  gear: GearIcon,
  bolt: BoltIcon,
  wallet: WalletIcon,
  car: CarIcon,
  phone: PhoneIcon,
  trophy: TrophyIcon,
  card: CardIcon,
};

export function ProductIcon({ id }: { id: IconId }) {
  const Icon = ICON_MAP[id];
  return <Icon />;
}

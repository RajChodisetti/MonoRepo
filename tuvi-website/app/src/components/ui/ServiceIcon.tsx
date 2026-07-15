export type ServiceIconName =
  | "website"
  | "app"
  | "ai"
  | "restaurant"
  | "data"
  | "growth"
  | "qr"
  | "rewards"
  | "voice"
  | "calendar"
  | "campaign"
  | "mail";

export default function ServiceIcon({
  name,
  className = "h-6 w-6",
}: {
  name: ServiceIconName;
  className?: string;
}) {
  return (
    <svg
      aria-hidden="true"
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.7"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      {name === "website" ? (
        <>
          <rect x="3" y="4" width="18" height="16" rx="2" />
          <path d="M3 8h18M7 6h.01M10 6h.01" />
        </>
      ) : null}
      {name === "app" ? (
        <>
          <rect x="7" y="2" width="10" height="20" rx="2.5" />
          <path d="M10 5h4M11 18h2" />
        </>
      ) : null}
      {name === "ai" ? (
        <>
          <path d="M12 3v3M5.6 5.6l2.1 2.1M3 12h3M18 12h3M16.3 7.7l2.1-2.1" />
          <rect x="6" y="7" width="12" height="12" rx="4" />
          <path d="M9 12h.01M15 12h.01M9.5 15c1.7 1.2 3.3 1.2 5 0" />
        </>
      ) : null}
      {name === "restaurant" ? (
        <>
          <path d="M4 3v7a3 3 0 0 0 3 3V3M7 13v8M14 3v8M14 7h5M19 3v18" />
        </>
      ) : null}
      {name === "data" ? (
        <>
          <path d="M4 20V10M10 20V4M16 20v-7M22 20H2" />
          <path d="m4 7 6-4 6 7 5-5" />
        </>
      ) : null}
      {name === "growth" ? (
        <>
          <path d="M4 16c4-1 7-4 8-8 3 2 5 5 6 9" />
          <path d="M5 19c4-2 8-2 14 0M12 8V3M9 5l3-2 3 2" />
        </>
      ) : null}
      {name === "qr" ? (
        <>
          <rect x="3" y="3" width="6" height="6" rx="1" />
          <rect x="15" y="3" width="6" height="6" rx="1" />
          <rect x="3" y="15" width="6" height="6" rx="1" />
          <path d="M15 15h2v2h-2zM19 15h2M19 19h2v2M15 19v2" />
        </>
      ) : null}
      {name === "rewards" ? (
        <>
          <path d="m12 3 2.7 5.5 6.1.9-4.4 4.3 1 6.1-5.4-2.9-5.4 2.9 1-6.1-4.4-4.3 6.1-.9L12 3Z" />
        </>
      ) : null}
      {name === "voice" ? (
        <>
          <rect x="9" y="3" width="6" height="11" rx="3" />
          <path d="M5 11a7 7 0 0 0 14 0M12 18v3M9 21h6" />
        </>
      ) : null}
      {name === "calendar" ? (
        <>
          <rect x="3" y="5" width="18" height="16" rx="2" />
          <path d="M8 3v4M16 3v4M3 10h18M8 15l2 2 5-5" />
        </>
      ) : null}
      {name === "campaign" ? (
        <>
          <path d="m4 13 12-5v8L4 13ZM16 10l3-2v8l-3-2M6 14l1 5h4l-2-6" />
        </>
      ) : null}
      {name === "mail" ? (
        <>
          <rect x="3" y="5" width="18" height="14" rx="2" />
          <path d="m4 7 8 6 8-6" />
        </>
      ) : null}
    </svg>
  );
}

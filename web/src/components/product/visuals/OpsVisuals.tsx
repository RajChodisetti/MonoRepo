import Image from "next/image";

/** Branded app — phone with menu */
export function BrandedAppPhoneVisual() {
  return (
    <div className="relative mx-auto w-[200px] sm:w-[220px]">
      <div className="overflow-hidden rounded-[32px] border-[6px] border-ink bg-bg shadow-[0_24px_50px_rgba(15,39,31,0.22)]">
        <div className="bg-parchment px-3 pb-4 pt-3">
          <div className="mx-auto mb-2 h-1 w-16 rounded-full bg-secondary/40" />
          <div className="relative h-[120px] overflow-hidden rounded-2xl">
            <Image src="/menu/birria-tacos.jpg" alt="" fill className="object-cover" sizes="200px" />
          </div>
          <p className="mt-3 text-[12px] font-bold text-ink">Your Restaurant</p>
          <div className="mt-2 space-y-2">
            <div className="h-2 w-[90%] rounded-full bg-sage" />
            <div className="h-2 w-[60%] rounded-full bg-sage" />
          </div>
          <div className="mt-3 rounded-xl bg-accent px-3 py-2 text-center text-[11px] font-semibold text-bg">
            Order ahead
          </div>
        </div>
      </div>
    </div>
  );
}

export function BrandedAppShowcaseVisual() {
  return (
    <div className="relative h-full min-h-[280px] w-full">
      <Image
        src="/image.png"
        alt=""
        fill
        className="object-contain object-bottom"
        sizes="(max-width: 1024px) 90vw, 560px"
      />
    </div>
  );
}

export function CampaignPromoVisual() {
  return (
    <div className="w-full max-w-[300px] rounded-[22px] bg-bg p-4 shadow-[0_16px_40px_rgba(15,39,31,0.14)]">
      <p className="text-[12px] font-semibold uppercase tracking-[0.12em] text-secondary">Campaign</p>
      <p className="mt-2 text-[18px] font-bold tracking-[-0.02em] text-ink">Taco Tuesday blast</p>
      <div className="mt-4 flex gap-2">
        {["Email", "SMS", "Push"].map((ch) => (
          <span
            key={ch}
            className="rounded-full bg-sage px-2.5 py-1 text-[11px] font-semibold text-ink"
          >
            {ch}
          </span>
        ))}
      </div>
      <div className="mt-4 rounded-xl bg-parchment p-3">
        <p className="text-[13px] font-semibold text-ink">2,480 guests reached</p>
        <p className="mt-1 text-[12px] text-muted">+18% orders this week</p>
      </div>
    </div>
  );
}

export function CampaignCalendarVisual() {
  return (
    <div className="grid w-full max-w-[280px] grid-cols-3 gap-2">
      {["Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].map((d, i) => (
        <div
          key={d}
          className={`rounded-2xl p-3 ${i === 1 ? "bg-accent text-bg" : "bg-bg text-ink shadow-sm"}`}
        >
          <p className="text-[11px] font-medium opacity-70">{d}</p>
          <p className="mt-1 text-[13px] font-bold">{i === 1 ? "Live" : "—"}</p>
        </div>
      ))}
    </div>
  );
}

export function EmailSmsPreviewVisual() {
  return (
    <div className="w-full max-w-[300px] space-y-3">
      <div className="rounded-2xl bg-bg p-4 shadow-[0_12px_32px_rgba(15,39,31,0.12)]">
        <p className="text-[11px] font-semibold text-secondary">Email</p>
        <p className="mt-1 text-[14px] font-bold text-ink">Come back for 15% off</p>
        <div className="mt-3 h-2 w-[85%] rounded-full bg-sage" />
        <div className="mt-2 h-2 w-[55%] rounded-full bg-sage" />
      </div>
      <div className="rounded-2xl bg-ink p-4 text-bg shadow-[0_12px_32px_rgba(15,39,31,0.18)]">
        <p className="text-[11px] font-semibold text-bg/70">SMS</p>
        <p className="mt-1 text-[14px] font-semibold leading-snug">
          Your loyalty reward is ready. Show this text at pickup tonight.
        </p>
      </div>
    </div>
  );
}

export function PushNotifStackVisual() {
  return (
    <div className="relative w-full max-w-[300px] space-y-2.5">
      {[
        { title: "Happy hour starts now", body: "20% off drinks until 7pm" },
        { title: "Your order is ready", body: "Pickup at the front counter" },
        { title: "Earn double points", body: "This weekend only" },
      ].map((n, i) => (
        <div
          key={n.title}
          className="rounded-2xl bg-bg px-4 py-3 shadow-[0_10px_28px_rgba(15,39,31,0.12)]"
          style={{ transform: `translateX(${i * 6}px)` }}
        >
          <div className="flex items-start gap-3">
            <span className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-accent text-[12px] font-bold text-bg">
              T
            </span>
            <div>
              <p className="text-[13px] font-bold text-ink">{n.title}</p>
              <p className="mt-0.5 text-[12px] text-muted">{n.body}</p>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

export function LoyaltyCardVisual() {
  return (
    <div className="w-full max-w-[280px] overflow-hidden rounded-[22px] bg-ink p-5 text-bg shadow-[0_16px_40px_rgba(15,39,31,0.2)]">
      <p className="text-[12px] font-medium text-bg/70">Loyalty</p>
      <p className="mt-1 text-[22px] font-bold tracking-[-0.03em]">240 points</p>
      <div className="mt-4 h-2 overflow-hidden rounded-full bg-white/15">
        <div className="h-full w-[68%] rounded-full bg-accent" />
      </div>
      <p className="mt-2 text-[12px] text-bg/75">60 points to your next free item</p>
      <div className="mt-5 rounded-xl bg-white/10 px-3 py-2.5 text-[13px] font-semibold">
        Free appetizer unlocked
      </div>
    </div>
  );
}

export function LoyaltyRewardsGridVisual() {
  return (
    <div className="grid w-full max-w-[300px] grid-cols-2 gap-2.5">
      {["Free drink", "10% off", "Dessert", "Double pts"].map((r) => (
        <div key={r} className="rounded-2xl bg-bg p-3.5 shadow-[0_8px_24px_rgba(15,39,31,0.1)]">
          <span className="flex h-8 w-8 items-center justify-center rounded-full bg-sage text-[14px] font-bold text-primary">
            ★
          </span>
          <p className="mt-2 text-[13px] font-semibold text-ink">{r}</p>
        </div>
      ))}
    </div>
  );
}

export function OwnerDashboardVisual() {
  return (
    <div className="w-full max-w-[340px] rounded-[22px] bg-bg p-4 shadow-[0_16px_40px_rgba(15,39,31,0.14)]">
      <p className="text-[13px] font-bold text-ink">Today overview</p>
      <div className="mt-3 grid grid-cols-2 gap-2.5">
        {[
          { label: "Orders", value: "86" },
          { label: "Sales", value: "$4.2k" },
          { label: "Avg ticket", value: "$48" },
          { label: "New guests", value: "19" },
        ].map((s) => (
          <div key={s.label} className="rounded-xl bg-parchment px-3 py-2.5">
            <p className="text-[11px] text-muted">{s.label}</p>
            <p className="mt-0.5 text-[18px] font-bold text-ink">{s.value}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

export function AnalyticsChartVisual() {
  return (
    <div className="w-full max-w-[320px] rounded-[22px] bg-bg p-4 shadow-[0_16px_40px_rgba(15,39,31,0.14)]">
      <p className="text-[13px] font-bold text-ink">Weekly sales</p>
      <div className="mt-4 flex h-[120px] items-end gap-2">
        {[42, 58, 51, 70, 63, 88, 76].map((h, i) => (
          <div key={i} className="flex flex-1 flex-col items-center gap-1.5">
            <div
              className="w-full rounded-t-md bg-accent"
              style={{ height: `${h}%`, opacity: i === 5 ? 1 : 0.55 }}
            />
            <span className="text-[9px] font-medium text-secondary">
              {"MTWTFSS"[i]}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

export function KitchenTicketVisual() {
  return (
    <div className="w-full max-w-[280px] rounded-[18px] bg-bg p-4 shadow-[0_16px_40px_rgba(15,39,31,0.14)] ring-1 ring-border">
      <div className="flex items-center justify-between">
        <p className="text-[12px] font-bold uppercase tracking-[0.1em] text-secondary">
          Ticket #184
        </p>
        <span className="rounded-full bg-accent/15 px-2 py-0.5 text-[11px] font-bold text-primary">
          New
        </span>
      </div>
      <p className="mt-2 text-[15px] font-bold text-ink">Pickup · ASAP</p>
      <ul className="mt-3 space-y-2 border-t border-border pt-3">
        {["1× Margherita Pizza", "1× Caesar Salad", "1× Garlic Bread"].map((item) => (
          <li key={item} className="text-[13px] font-medium text-ink">
            {item}
          </li>
        ))}
      </ul>
    </div>
  );
}

export function PosSyncVisual() {
  return (
    <div className="w-full max-w-[300px] rounded-[22px] bg-bg p-4 shadow-[0_16px_40px_rgba(15,39,31,0.14)]">
      <div className="flex items-center gap-2">
        <span className="h-2.5 w-2.5 rounded-full bg-accent" />
        <p className="text-[14px] font-semibold text-ink">POS synced</p>
      </div>
      <div className="mt-4 space-y-2.5">
        {["Toast", "Square", "Clover"].map((name) => (
          <div
            key={name}
            className="flex items-center justify-between rounded-xl bg-parchment px-3 py-2.5"
          >
            <span className="text-[13px] font-semibold text-ink">{name}</span>
            <span className="rounded-full bg-accent/15 px-2 py-0.5 text-[10px] font-bold text-primary">
              Connected
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

export function OpsPhotoFillVisual({ src }: { src: string }) {
  return (
    <div className="absolute inset-0">
      <Image src={src} alt="" fill className="object-cover" sizes="(max-width: 1024px) 90vw, 560px" />
      <div
        className="absolute inset-0 bg-gradient-to-t from-ink/50 via-transparent to-transparent"
        aria-hidden="true"
      />
    </div>
  );
}

export function AppPhotoFillVisual() {
  return <OpsPhotoFillVisual src="/fruits.png" />;
}

export function KitchenPhotoFillVisual() {
  return <OpsPhotoFillVisual src="/guides/interview.jpg" />;
}

export function OwnerPhotoFillVisual() {
  return <OpsPhotoFillVisual src="/resources/resource-help-hero.png" />;
}

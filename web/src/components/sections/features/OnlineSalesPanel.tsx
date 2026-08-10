"use client";

import Image from "next/image";
import { useEffect, useState } from "react";

const POINTS_PER_ADD = 156;
const REWARD_START = 4200;

type MenuItem = {
  id: string;
  name: string;
  price: string;
  badge?: string;
  imageUrl: string;
};

const MENU: MenuItem[] = [
  {
    id: "caesar",
    name: "Caesar Salad",
    price: "$11.99",
    badge: "New",
    imageUrl: "/menu/caesar-salad.jpg",
  },
  {
    id: "garlic",
    name: "Garlic Bread",
    price: "$7.99",
    imageUrl: "/menu/garlic-bread-v2.jpg",
  },
  {
    id: "pizza",
    name: "Pepperoni Pizza",
    price: "$14.99",
    imageUrl: "/menu/pepperoni-pizza.jpg",
  },
  {
    id: "pasta",
    name: "Pasta Alfredo",
    price: "$13.49",
    imageUrl: "/menu/pasta-alfredo.jpg",
  },
  {
    id: "burger",
    name: "Classic Burger",
    price: "$12.99",
    imageUrl: "/menu/classic-burger.jpg",
  },
];

function CartIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-[17px] w-[17px] shrink-0 text-[#1a1a1a]" aria-hidden="true">
      <path
        fill="currentColor"
        d="M7 18a2 2 0 1 0 0 4 2 2 0 0 0 0-4Zm10 0a2 2 0 1 0 0 4 2 2 0 0 0 0-4ZM7.2 6h13.1l-1.2 6.2a2 2 0 0 1-2 1.6H9.1a2 2 0 0 1-2-1.6L5.6 3.5A1.5 1.5 0 0 0 4.1 2H2v1.8h1.7l1.5 7.7a3.8 3.8 0 0 0 3.7 3H17v-1.8H9a2 2 0 0 1-2-1.6L7.2 6Z"
      />
    </svg>
  );
}

function PlusIcon() {
  return (
    <svg viewBox="0 0 20 20" className="h-5 w-5" aria-hidden="true">
      <path d="M10 4v12M4 10h12" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg viewBox="0 0 20 20" className="h-5 w-5 text-white" aria-hidden="true">
      <path
        d="M4.5 10.5 8 14l7.5-8"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function OdometerDigit({ digit, height = 20 }: { digit: string; height?: number }) {
  if (digit === ",") {
    return (
      <span className="inline-flex w-[0.32em] items-end justify-center text-[#1a1a1a]" style={{ height }}>
        ,
      </span>
    );
  }
  const value = Number(digit);
  return (
    <span className="relative inline-block overflow-hidden text-[#1a1a1a]" style={{ height, width: "0.6em" }}>
      <span
        className="absolute left-0 top-0 flex w-full flex-col transition-transform duration-500 ease-[cubic-bezier(0.22,1,0.36,1)]"
        style={{ transform: `translateY(-${value * height}px)` }}
      >
        {Array.from({ length: 10 }, (_, n) => (
          <span key={n} className="flex items-center justify-center font-semibold tabular-nums" style={{ height }}>
            {n}
          </span>
        ))}
      </span>
    </span>
  );
}

function OdometerNumber({ value }: { value: number }) {
  const formatted = Math.round(value).toLocaleString("en-US");
  return (
    <span className="inline-flex items-center text-[14px] font-semibold tracking-[-0.02em] text-[#1a1a1a] sm:text-[15px]">
      {formatted.split("").map((char, index) => (
        <OdometerDigit key={index} digit={char} />
      ))}
    </span>
  );
}

function MenuCard({
  item,
  emphasized,
  ticked,
}: {
  item: MenuItem;
  emphasized: boolean;
  ticked: boolean;
}) {
  return (
    <div
      className={`flex h-[88px] w-full items-center gap-3 rounded-2xl bg-white p-3 ring-1 transition-shadow duration-300 ${
        emphasized ? "shadow-[0_8px_24px_rgba(0,0,0,0.06)] ring-black/10" : "ring-black/[0.05]"
      }`}
    >
      <div className="relative h-[62px] w-[62px] shrink-0 overflow-hidden rounded-xl bg-[#ebe7e2]">
        <Image src={item.imageUrl} alt="" fill className="object-cover" sizes="62px" />
        {item.badge ? (
          <span className="absolute left-1.5 top-1.5 rounded-full bg-[#9fd4a3] px-1.5 py-0.5 text-[9px] font-semibold text-white">
            {item.badge}
          </span>
        ) : null}
      </div>

      <div className="min-w-0 flex-1">
        <p
          className={`truncate text-[14px] tracking-[-0.01em] ${
            emphasized || ticked ? "font-semibold text-[#1a1a1a]" : "font-medium text-[#8a8580]"
          }`}
        >
          {item.name}
        </p>
        <div className="mt-2 space-y-1.5">
          <div className="h-1.5 w-[78%] rounded-full bg-[#ebe7e2]" />
          <div className="h-1.5 w-[52%] rounded-full bg-[#ebe7e2]" />
        </div>
        <p className="mt-2 text-[12px] text-[#9a9590]">{item.price}</p>
      </div>

      <span
        className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl transition-all duration-300 ${
          ticked
            ? "bg-[#1f9a3f] text-white shadow-[0_6px_16px_rgba(31,154,63,0.35)]"
            : "bg-white text-[#9a9590] ring-1 ring-black/10"
        }`}
      >
        {ticked ? <CheckIcon /> : <PlusIcon />}
      </span>
    </div>
  );
}

type Phase = "idle" | "tick" | "reward" | "exit";

type OnlineSalesPanelProps = {
  onProgress?: (progress: number) => void;
  onComplete?: () => void;
};

export default function OnlineSalesPanel({ onProgress, onComplete }: OnlineSalesPanelProps) {
  const [top, setTop] = useState<MenuItem>(MENU[0]);
  const [bottom, setBottom] = useState<MenuItem>(MENU[1]);
  const [incoming, setIncoming] = useState<MenuItem>(MENU[2]);
  const [phase, setPhase] = useState<Phase>("idle");
  const [snap, setSnap] = useState(false); // disable CSS transition when resetting after exit
  const [points, setPoints] = useState(REWARD_START);
  const [badge, setBadge] = useState(4);
  const [badgePulse, setBadgePulse] = useState(false);

  useEffect(() => {
    let cancelled = false;
    const timers: number[] = [];
    let raf = 0;

    const report = (value: number) => {
      if (!cancelled) onProgress?.(Math.min(1, Math.max(0, value)));
    };

    const wait = (ms: number) =>
      new Promise<void>((resolve) => {
        timers.push(window.setTimeout(resolve, ms));
      });

    const animateCount = (from: number, to: number, duration: number, p0: number, p1: number) =>
      new Promise<void>((resolve) => {
        const start = performance.now();
        const tick = (now: number) => {
          if (cancelled) {
            resolve();
            return;
          }
          const t = Math.min(1, (now - start) / duration);
          const eased = 1 - Math.pow(1 - t, 3);
          setPoints(from + (to - from) * eased);
          report(p0 + (p1 - p0) * eased);
          if (t < 1) raf = requestAnimationFrame(tick);
          else {
            setPoints(to);
            report(p1);
            resolve();
          }
        };
        raf = requestAnimationFrame(tick);
      });

    const ITEMS_PER_CYCLE = 3;

    const run = async () => {
      let localTop = MENU[0];
      let localBottom = MENU[1];
      let localIncoming = MENU[2];
      let ptr = 3;
      let localPoints = REWARD_START;
      let localBadge = 4;

      setTop(localTop);
      setBottom(localBottom);
      setIncoming(localIncoming);
      setPhase("idle");
      setPoints(REWARD_START);
      setBadge(4);
      setBadgePulse(false);
      report(0);

      for (let item = 0; item < ITEMS_PER_CYCLE; item += 1) {
        if (cancelled) return;

        const base = item / ITEMS_PER_CYCLE;
        const span = 1 / ITEMS_PER_CYCLE;

        setPhase("idle");
        report(base + span * 0.1);
        await wait(850);
        if (cancelled) return;

        setPhase("tick");
        report(base + span * 0.28);
        await wait(450);
        if (cancelled) return;

        setPhase("reward");
        setBadgePulse(true);
        localBadge += 1;
        setBadge(localBadge);
        const nextPoints = localPoints + POINTS_PER_ADD;
        await animateCount(localPoints, nextPoints, 1000, base + span * 0.35, base + span * 0.75);
        localPoints = nextPoints;
        if (cancelled) return;

        await wait(500);
        if (cancelled) return;

        setPhase("exit");
        report(base + span * 0.9);
        await wait(700);
        if (cancelled) return;

        const newIncoming = MENU[ptr % MENU.length];
        localBottom = localTop;
        localTop = localIncoming;
        localIncoming = newIncoming;
        ptr += 1;

        setSnap(true);
        setBottom(localBottom);
        setTop(localTop);
        setIncoming(localIncoming);
        setPhase("idle");
        setBadgePulse(false);
        report(base + span);
        await wait(40);
        if (cancelled) return;
        setSnap(false);
        await wait(160);
      }

      if (cancelled) return;
      report(1);
      await wait(500);
      if (cancelled) return;
      onComplete?.();
    };

    void run();

    return () => {
      cancelled = true;
      timers.forEach((id) => window.clearTimeout(id));
      cancelAnimationFrame(raf);
      onProgress?.(0);
    };
  }, [onProgress, onComplete]);

  const isExiting = phase === "exit";
  const isTicked = phase === "tick" || phase === "reward";
  const moveClass = snap
    ? "absolute inset-x-0 transition-none"
    : "absolute inset-x-0 transition-all duration-700 ease-[cubic-bezier(0.22,1,0.36,1)]";

  return (
    <div
      className="grid min-h-[520px] items-center gap-10 rounded-[28px] px-6 py-16 sm:min-h-[560px] sm:rounded-[32px] sm:px-10 sm:py-20 lg:grid-cols-2 lg:gap-12 lg:px-14 lg:py-24"
      style={{ backgroundColor: "#f2ecdf" }}
    >
      <div>
        <p className="text-[14px] font-medium text-[#7a7268] sm:text-[15px]">Commission-free ordering</p>
        <h3 className="mt-3 text-[clamp(1.85rem,3.2vw,2.85rem)] font-bold leading-[1.12] tracking-[-0.04em] text-[#1a1a1a]">
          <span className="block">Let guests order on</span>
          <span className="block">your site — and keep</span>
          <span className="block">every dollar you earn</span>
        </h3>
      </div>

      <div className="mx-auto flex w-full max-w-[380px] flex-col gap-3">
        <div className="flex justify-center">
          <div className="inline-flex items-center gap-2.5 rounded-full border border-[#1a1a1a]/10 bg-[#fff8ee] px-4 py-2.5 shadow-[0_10px_28px_rgba(26,26,26,0.08)]">
            <CartIcon />
            <p className="flex items-center gap-1.5 whitespace-nowrap text-[14px] font-medium text-[#1a1a1a]">
              <span>You&apos;ll earn</span>
              <OdometerNumber value={points} />
              <span>points</span>
            </p>
            <span
              className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-[12px] font-bold text-[#fff8ee] transition-transform duration-300 ${
                badgePulse ? "scale-110" : "scale-100"
              }`}
              style={{ backgroundColor: "#c45c26" }}
            >
              {badge}
            </span>
          </div>
        </div>

        <div className="relative h-[196px] overflow-hidden">
          {/* Incoming — enters from top on exit */}
          <div
            className={`${moveClass} ${
              isExiting ? "top-0 translate-y-0 opacity-100" : "top-0 -translate-y-[110%] opacity-0"
            }`}
          >
            <MenuCard key={incoming.id} item={incoming} emphasized={false} ticked={false} />
          </div>

          {/* Top waiting card — moves down to bottom on exit */}
          <div className={`${moveClass} ${isExiting ? "top-[100px]" : "top-0"}`}>
            <MenuCard key={top.id} item={top} emphasized={false} ticked={false} />
          </div>

          {/* Bottom active — ticks, then exits downward */}
          <div
            className={`${moveClass} ${
              isExiting ? "top-[210px] opacity-0" : "top-[100px] opacity-100"
            }`}
          >
            <MenuCard key={bottom.id} item={bottom} emphasized={!isExiting} ticked={isTicked} />
          </div>
        </div>
      </div>
    </div>
  );
}

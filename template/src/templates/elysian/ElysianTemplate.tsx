"use client";

import { useCallback, useRef, useState } from "react";
import type { RestaurantContent } from "@/data/types/restaurant";
import { mapElysianContent } from "./lib/mapContent";
import { useRevealObserver, useScrollChrome, useCounterAnimation } from "./hooks/useElysianEffects";
import { useThemeToggle } from "./hooks/useThemeToggle";
import ElysianLoader from "./components/ElysianLoader";
import ElysianCursor from "./components/ElysianCursor";
import ElysianScrollProgress from "./components/ElysianScrollProgress";
import ElysianNav from "./components/ElysianNav";
import ElysianHero from "./components/ElysianHero";
import ElysianAbout from "./components/ElysianAbout";
import ElysianDishes from "./components/ElysianDishes";
import ElysianMenu from "./components/ElysianMenu";
import ElysianTestimonials from "./components/ElysianTestimonials";
import ElysianGallery from "./components/ElysianGallery";
import ElysianReservation from "./components/ElysianReservation";
import ElysianWhy from "./components/ElysianWhy";
import ElysianStats from "./components/ElysianStats";
import ElysianFAQ from "./components/ElysianFAQ";
import ElysianContact from "./components/ElysianContact";
import ElysianFooter from "./components/ElysianFooter";
import ElysianScrollTop from "./components/ElysianScrollTop";
import "./theme.css";

export default function ElysianTemplate({ restaurant }: { restaurant: RestaurantContent }) {
  const content = mapElysianContent(restaurant);
  const rootRef = useRef<HTMLDivElement>(null);
  const [heroLoaded, setHeroLoaded] = useState(false);
  const [progress, setProgress] = useState(0);
  const [navScrolled, setNavScrolled] = useState(false);
  const [scrollTopVisible, setScrollTopVisible] = useState(false);
  const { theme, toggle: toggleTheme } = useThemeToggle();

  const onScroll = useCallback((scrollTop: number, docHeight: number) => {
    setProgress(docHeight > 0 ? (scrollTop / docHeight) * 100 : 0);
    setNavScrolled(scrollTop > 40);
    setScrollTopVisible(scrollTop > 600);
  }, []);

  useScrollChrome(onScroll);
  useRevealObserver(rootRef);
  useCounterAnimation(rootRef);

  return (
    <div
      ref={rootRef}
      className="elysian-root"
      data-theme={theme}
      style={{
        fontFamily: "var(--font-elysian-body), Poppins, sans-serif",
      }}
    >
      <ElysianLoader name={restaurant.name} onDone={() => setHeroLoaded(true)} />
      <ElysianCursor />
      <ElysianScrollProgress width={progress} />
      <ElysianNav
        name={content.hero.name}
        nameAccent={content.hero.nameAccent}
        scrolled={navScrolled}
        theme={theme}
        onToggleTheme={toggleTheme}
        showDishes={content.show.dishes}
      />

      <main id="home">
        <ElysianHero hero={content.hero} loaded={heroLoaded} />
        <ElysianAbout about={content.about} />
        {content.show.dishes ? <ElysianDishes dishes={content.dishes} /> : null}
        {content.show.menu ? (
          <ElysianMenu menuItems={content.menuItems} menuTabs={content.menuTabs} />
        ) : null}
        {content.show.testimonials ? <ElysianTestimonials reviews={content.reviews} /> : null}
        {content.show.gallery ? <ElysianGallery images={content.gallery} /> : null}
        <ElysianReservation restaurant={restaurant} />
        {content.show.why ? <ElysianWhy cards={content.experienceCards} /> : null}
        {content.show.stats ? <ElysianStats stats={content.stats} /> : null}
        {content.show.faq ? <ElysianFAQ items={content.faq} /> : null}
        <ElysianContact contact={content.contact} name={restaurant.name} showMap={content.show.map} />
      </main>

      <ElysianFooter
        name={content.hero.name}
        nameAccent={content.hero.nameAccent}
        footer={content.footer}
        showInsta={content.show.insta}
      />
      <ElysianScrollTop visible={scrollTopVisible} />
    </div>
  );
}

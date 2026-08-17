import type { RestaurantContent } from "@/data/types/restaurant";
import Navigation from "./components/Navigation";
import MobileStickyBar from "./components/MobileStickyBar";
import HeroRestaurantVideo from "./components/HeroRestaurantVideo";
import ScrollVideoSection from "./components/ScrollVideoSection";
import RestaurantStorySection from "./components/RestaurantStorySection";
import ScrollDishSlideshow from "./components/ScrollDishSlideshow";
import MenuPreview from "./components/MenuPreview";
import AtmosphereGallery from "./components/AtmosphereGallery";
import ReservationCTA from "./components/ReservationCTA";
import ReviewsSection from "./components/ReviewsSection";
import LocationHours from "./components/LocationHours";
import FinalCTA from "./components/FinalCTA";
import Footer from "./components/Footer";
import { buildRestaurantJsonLd } from "./seo";

export default function CinematicTemplate({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  const jsonLd = buildRestaurantJsonLd(restaurant);

  return (
    <div data-template="1">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <Navigation restaurant={restaurant} />
      <main>
        <HeroRestaurantVideo restaurant={restaurant} />
        <ScrollVideoSection
          videoSrc={restaurant.videos.kitchen.src}
          posterSrc={restaurant.videos.kitchen.poster}
          posterMedia={restaurant.heroMedia}
          eyebrow="In the kitchen"
          title="Where every plate begins with craft"
          description={`At ${restaurant.name}, our kitchen moves with rhythm — fire, seasoning, and the patience to get it right.`}
          ctaLabel="Explore Menu"
          ctaHref="#menu"
        />
        <RestaurantStorySection
          steps={restaurant.storySteps}
          restaurantName={restaurant.name}
        />
        <ScrollDishSlideshow dishes={restaurant.signatureDishes} />
        <MenuPreview restaurant={restaurant} />
        <AtmosphereGallery images={restaurant.galleryImages} />
        <ReservationCTA restaurant={restaurant} />
        <ReviewsSection
          reviews={restaurant.reviews}
          rating={restaurant.rating}
          reviewsCount={restaurant.reviewsCount}
        />
        <LocationHours restaurant={restaurant} />
        <FinalCTA restaurant={restaurant} />
      </main>
      <Footer restaurant={restaurant} />
      <MobileStickyBar restaurant={restaurant} />
    </div>
  );
}

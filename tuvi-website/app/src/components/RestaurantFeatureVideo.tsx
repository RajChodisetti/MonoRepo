type RestaurantFeatureVideoProps = {
  poster: string;
  src: string;
  title: string;
};

export default function RestaurantFeatureVideo({ poster, src, title }: RestaurantFeatureVideoProps) {
  return (
    <div className="relative aspect-video w-full overflow-hidden bg-ink">
      <video
        aria-label={`${title} product demo`}
        autoPlay
        className="pointer-events-none block h-full w-full select-none object-cover"
        controls={false}
        controlsList="nodownload nofullscreen noplaybackrate"
        disablePictureInPicture
        src={src}
        poster={poster}
        loop
        muted
        playsInline
        preload="auto"
      >
        Your browser does not support embedded videos.
      </video>
    </div>
  );
}

"use client";

import Image from "next/image";
import { forwardRef, type CSSProperties, type MouseEventHandler, type Ref } from "react";

type ElysianImageProps = {
  src: string;
  alt: string;
  className?: string;
  style?: CSSProperties;
  width?: number;
  height?: number;
  fill?: boolean;
  sizes?: string;
  priority?: boolean;
  id?: string;
  onClick?: () => void;
  onMouseMove?: MouseEventHandler<HTMLDivElement>;
};

/** Remote Google / CDN photos via Next.js image optimizer (fixes broken plain img tags). */
const ElysianImage = forwardRef(function ElysianImage(
  {
    src,
    alt,
    className,
    style,
    width,
    height,
    fill,
    sizes,
    priority,
    id,
    onClick,
    onMouseMove,
  }: ElysianImageProps,
  ref: Ref<HTMLDivElement>,
) {
  if (fill) {
    return (
      <div
        ref={ref}
        className={className}
        style={{ position: "relative", width: "100%", height: "100%", ...style }}
        id={id}
        onClick={onClick}
        onMouseMove={onMouseMove}
      >
        <Image
          src={src}
          alt={alt}
          fill
          className="object-cover"
          sizes={sizes || "100vw"}
          priority={priority}
        />
      </div>
    );
  }

  return (
    <Image
      src={src}
      alt={alt}
      width={width!}
      height={height!}
      className={className}
      style={style}
      sizes={sizes}
      priority={priority}
      id={id}
      onClick={onClick}
    />
  );
});

export default ElysianImage;

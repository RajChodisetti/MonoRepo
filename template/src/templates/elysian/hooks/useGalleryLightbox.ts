"use client";

import { useCallback, useState } from "react";

export function useGalleryLightbox() {
  const [open, setOpen] = useState(false);
  const [src, setSrc] = useState("");
  const [alt, setAlt] = useState("");

  const openImage = useCallback((url: string, imageAlt: string) => {
    setSrc(url.replace(/w=\d+/, "w=1600"));
    setAlt(imageAlt);
    setOpen(true);
  }, []);

  const close = useCallback(() => setOpen(false), []);

  return { open, src, alt, openImage, close };
}

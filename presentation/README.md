# Tuvi Presentation Site

Animated HTML presentation showcasing Tuvi services.

## Open locally

```bash
cd MonoRepo/presentation
python3 -m http.server 5500
```

Then visit: **http://localhost:5500**

Or simply open `index.html` in your browser.

## Add your demo video

Edit `index.html` — find `#video-container` and replace the placeholder with:

```html
<video src="assets/demo.mp4" controls playsinline poster="assets/poster.jpg"></video>
```

Or YouTube:

```html
<iframe src="https://www.youtube.com/embed/YOUR_VIDEO_ID" allowfullscreen></iframe>
```

Put video files in `assets/` folder.

## Custom images

Current images are from Unsplash (free to use). To replace with AI-generated images, see **IMAGE_PROMPTS.md** for ready-to-use prompts.

## Sections

1. Hero — platform overview
2. Services — 8 service cards (websites, QR ordering, rewards, voice AI, etc.)
3. Custom Apps — QR ordering & rewards membership deep-dive
4. Voice AI — feature breakdown
4. Video demo — empty container for your video
5. Templates gallery — Cinematic & Aurora
6. Process — 4-step workflow
7. CTA — contact / book demo

# Restaurant Demo Website Template

Parameterized HTML/CSS/JS template — builds a full restaurant website from `data/restaurants_data.json`.

## Quick start

From **MonoRepo root** (required — browser needs HTTP for JSON fetch):

```bash
cd MonoRepo
python3 -m http.server 8080
```

Open in browser:

| URL | Restaurant |
|-----|------------|
| http://localhost:8080/template/legacy/index.html?id=0 | First (`restaurants[0]`) — Bistro Moncur |
| http://localhost:8080/template/legacy/index.html?id=1 | Second (`restaurants[1]`) |
| http://localhost:8080/template/index.html?id=2 | Third (`restaurants[2]`) |

## Change default restaurant

Edit `template/js/config.js`:

```js
window.RESTAURANT_CONFIG = {
  index: 1,   // restaurants[1]
  dataPath: "../data/restaurants_data.json",
};
```

Or use URL query: `?id=1` or `?index=1`

## Data mapping

| JSON field | Template section |
|------------|------------------|
| `name`, `cuisines`, `rating`, `price_level` | Hero |
| `location`, `cuisines` | About |
| `menu_items[]` | Menu (grouped by category) |
| `images.gallery`, menu photos | Gallery |
| `reviews[]` | Reviews |
| `hours`, `location`, `contact` | Visit + Map |
| `contact.phone/email/website` | Reserve CTA |

## Files

```
template/
  index.html      Page structure
  css/styles.css  Premium dark dining UI
  js/config.js    Default restaurant index + data path
  js/app.js       Data loader & renderer
```

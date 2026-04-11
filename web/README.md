# PhysicsCopilot — Landing Page

Static HTML/CSS landing page for [physicscopilot.app](https://physicscopilot.app).

## Stack

- Plain HTML5 + CSS3 (no framework, no build step)
- Vanilla JavaScript for scroll animations and the contact form
- Hosted as static files (Vercel / Cloudflare Pages / any static host)

## Files

| File | Purpose |
|------|---------|
| `index.html` | Main landing page (hero, features, pricing, contact) |
| `og-image.png` | Open Graph image for social sharing previews (1200×630) |

## Local development

Open `index.html` directly in a browser — no server or build step required:

```bash
# macOS / Linux
open web/index.html

# Windows
start web/index.html
```

Or serve with any static server:

```bash
npx serve web/
```

## Contact form

The contact form in the Enterprise section uses [Formspree](https://formspree.io).
The endpoint is configured in `index.html` (`action="https://formspree.io/f/xpwzgkql"`).
To use a different Formspree project, replace the form ID in that attribute.

## Deployment

The page deploys automatically on push to `main` via the CI/CD pipeline.
No build step needed — the `web/` directory is served as-is.

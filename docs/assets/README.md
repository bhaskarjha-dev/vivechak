# Vivechak Brand Assets

This directory contains the official production vector assets for **Vivechak (विवेचक)**.

For complete geometric construction blueprints, mathematical verification metrics, optical sizing rules, and color tokens, see [**`docs/BRAND.md`**](../BRAND.md).

---

## Asset Index

| File | Purpose | Colorway | Recommended Surface |
|---|---|---|---|
| [`logo.svg`](logo.svg) | **Primary Brand Mark** (512×512) | Ink Indigo (`#12203F`) + Saffron (`#F2A33A`) | White & Light backgrounds (`#FFFFFF`, `#F8FAFC`) |
| [`logo-dark.svg`](logo-dark.svg) | **Reversed Brand Mark** (512×512) | Warm Paper (`#F7F4EC`) + Saffron (`#F2A33A`) | Dark backgrounds (`#0D1117`, `#12203F`) |
| [`logo-mono.svg`](logo-mono.svg) | **Monochrome Mark** (512×512) | Carbon Black (`#000000`) | Single-color print & monochrome documents |
| [`favicon.svg`](favicon.svg) | **Optical Cut** (32×32) | Ink Indigo (`#12203F`) + Saffron (`#F2A33A`) | Browser favicons, IDE tabs, small app shortcuts |

---

## Quick Usage

### Theme-Adaptive GitHub / Markdown Embedding
```html
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo-dark.svg">
    <img src="docs/assets/logo.svg" alt="Vivechak Logo" width="180" />
  </picture>
</p>
```

### HTML Head Favicon
```html
<link rel="icon" type="image/svg+xml" href="/docs/assets/favicon.svg" />
```

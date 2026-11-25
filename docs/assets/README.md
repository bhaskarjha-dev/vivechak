# Vivechak Brand & Diagram Assets

This directory contains the official production vector and raster assets for **Vivechak (विवेचक)**.

For complete geometric construction blueprints, mathematical verification metrics, optical sizing rules, and color tokens, see [**`docs/BRAND.md`**](../BRAND.md).

---

## Asset Index

| File | Purpose | Colorway | Recommended Surface |
|---|---|---|---|
| [`logo.svg`](logo.svg) | **Primary Brand Mark** (512×512) | Ink Indigo (`#12203F`) + Saffron (`#F2A33A`) | White & Light backgrounds (`#FFFFFF`, `#F8FAFC`) |
| [`logo-dark.svg`](logo-dark.svg) | **Reversed Brand Mark** (512×512) | Warm Paper (`#F7F4EC`) + Saffron (`#F2A33A`) | Dark backgrounds (`#0D1117`, `#12203F`) |
| [`logo-mono.svg`](logo-mono.svg) | **Monochrome Mark** (512×512) | Carbon Black (`#000000`) | Single-color print & monochrome documents |
| [`favicon.svg`](favicon.svg) | **Optical Cut** (32×32) | Ink Indigo (`#12203F`) + Saffron (`#F2A33A`) | Browser favicons, IDE tabs, small app shortcuts |
| [`wordmark-light.svg`](wordmark-light.svg) | **Devanagari Vector Wordmark (Light)** | Ink Indigo (`#12203F`) | Pure-vector Devanagari script for light surfaces (zero OS font dependency) |
| [`wordmark-dark.svg`](wordmark-dark.svg) | **Devanagari Vector Wordmark (Dark)** | Warm Paper (`#F7F4EC`) | Pure-vector Devanagari script for dark surfaces (zero OS font dependency) |
| [`workflow-light.svg`](workflow-light.svg) | **5-Stage Lifecycle Diagram (Light)** (1080×264) | Slate Paper (`#F8FAFC`) + Ink Indigo + Saffron | Light-theme GitHub & documentation surfaces |
| [`workflow-dark.svg`](workflow-dark.svg) | **5-Stage Lifecycle Diagram (Dark)** (1080×264) | Deep Navy (`#0B1120`) + Warm Paper + Saffron | Dark-theme GitHub & documentation surfaces |
| [`social-preview.png`](social-preview.png) | **OpenGraph Banner** (1280×640) | Ink Indigo Canvas + Saffron Accent | GitHub Social Preview, X/Twitter cards, LinkedIn |
| [`social-preview.svg`](social-preview.svg) | **Vector Master** (1280×640) | Ink Indigo Canvas + Saffron Accent | Pure-vector high-DPI master (zero OS font dependency) |

---

## Quick Usage

### Theme-Adaptive GitHub / Markdown Embedding
```html
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo-dark.svg">
    <img src="docs/assets/logo.svg" alt="Vivechak" width="84" height="84">
  </picture>
</p>
```

### HTML Head Favicon
```html
<link rel="icon" type="image/svg+xml" href="/docs/assets/favicon.svg" />
```

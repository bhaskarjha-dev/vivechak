# Vivechak (विवेचक) — Brand Identity & Design System Specification
### Mathematical Geometry, Optical Sizing & Design Tokens

---

## 1. Philosophical & Conceptual Foundations

**Vivechak (विवेचक)** is rooted in classical Sanskrit epistemological grammar:
- **vi-** (वि): prefix denoting *"apart"*, *"distinctly"*, *"without confusion"*
- **√vic** (विच्): verbal root meaning *"to sift"*, *"to separate"*, *"to discern truth from appearance"*
- **-aka** (अक): agentive suffix denoting *"the agent who performs the action"*

Literally, **Vivechak** is *"the discerning analyst"* or *"the epistemic sieve"*.

```
                            S = 36 (Stroke Unit)
      ┌─────────────────────────────────────────────────────────────┐
      │                      9S Shirorekha Bar                      │  r = S/2
      └──────────────┬───────────────────────────────┬──────────────┘
                      \                             /
                       \      Equilateral Hole     /   30° off vertical
                        \        Side = 4S        /
                         \                       /
                          \                     /
                           \                   /
                            \                 /
                             \               /
                              \             /
                               \           /
                                ─────┬─────   Tip r = S/2
                                     │ Gap = 0.6S = 21.6
                                     ●   Saffron Dot (Ø 1.2S = 43.2)
```

The brand mark fuses two intellectual traditions:
1. **The Shirorekha (शिरोरेखा) + Latin 'V':** The horizontal headline bar that unifies Devanagari script is seamlessly fused with the Latin letter 'V'. This symbolizes the intersection of Indian epistemic inquiry (*Nyāya* and *Pramāṇa-śāstra*) with modern computational software architecture.
2. **The Funnel / Sieve:** The inverted triangular taper serves as a visual metaphor for **Principle P1** (Context Architecture Law) and **Principle P8** (Structured Falsification): voluminous, noisy LLM recall and ungrounded architectural assertions are poured into the funnel.
3. **The Saffron Truth Drop (`#F2A33A`):** Suspended cleanly below the tip of the funnel is a single spherical drop. It represents the verified, corroborated, evidence-graded claim that survives rigorous falsification to form the bedrock of the Founding Architecture Document (FAD).

---

## 2. Construction of the Primary Mark (512×512 Canvas)

The canonical mark is constructed with exact parametric ratios based on the fundamental stroke unit **$S = 36\text{ px}$** on a **512 × 512** coordinate canvas.

```
Bounding Box: 9S × 7.76S (324 × 279.5 px), centred on 512 × 512
Stroke Unit: S = 36 px
```

### Parametric Construction Values ($S = 36$)
- **Headline Bar (Shirorekha):** $9S \times S$ ($324 \times 36\text{ px}$), pill ends $r = S/2 = 18\text{ px}$. Coordinates: $X \in [94, 418]$, $Y \in [116.25, 152.25]$.
- **Funnel Arms:** $S = 36\text{ px}$ uniform thickness; every edge angles exactly $30^\circ$ off vertical.
- **Inner Cutout (Hole):** Equilateral triangle with side $4S = 144\text{ px}$, top edge flush with the underside of the bar ($Y = 152.25$), rounded corners $r = 0.2S = 7.2\text{ px}$.
- **Outer Tip Join:** Round join with radius $r = S/2 = 18\text{ px}$.
- **Bar & Funnel Transition:** Bar underside meets V arms in smooth concave fillets of radius $r = 0.3S = 10.8\text{ px}$, identical on both sides.
- **Sifted Truth Drop:** Diameter $\varnothing\ 1.2S = 43.2\text{ px}$ (radius $r = 21.6\text{ px}$), center at $(256, 374.16)$.
- **Tip-to-Drop Clearance Gap:** Uniform negative space of $0.6S = 21.6\text{ px}$ from the lowest point of the V tip ($Y = 330.96$) to the top of the drop ($Y = 352.56$).

### Measured Verification on Delivered Vector File
The production SVG was verified against strict geometric tolerances:
- **Stroke Width:** Bar $36.000\text{ px}$, V arms $36.003\text{ px}$ and $35.996\text{ px}$.
- **Edge Angles:** $29.999^\circ$ outer and $29.996^\circ$ inner, measured from vertical.
- **Equilateral Hole:** Side lengths $143.99\text{ px}$, $144.01\text{ px}$, $144.01\text{ px}$; internal angles $60.004^\circ$, $60.004^\circ$, $59.991^\circ$.
- **Clearance Gap:** $21.597\text{ px}$ (target $0.6S = 21.600\text{ px}$).
- **Drop Diameter:** $\varnothing\ 43.20\text{ px} \times 43.19\text{ px}$ on the render.
- **Mirror Symmetry:** Vertex error $0.0\text{ px}$. Flip-and-diff difference of at most $2.8 / 255$ after supersampling.
- **Tangent Continuity:** Sharpest join has a $0.07^\circ$ tangent break, ensuring completely smooth curves.

---

## 3. Optical Sizing Specification: Favicon Cut ($S = 4$ on 32×32 Grid)

Directly scaling a 512px vector to 16px or 32px causes severe visual degradation: the 7% stroke weight antialiases into a blurry haze, the hole closes up, and the small saffron dot vanishes.

The **Favicon Cut** ([`docs/assets/favicon.svg`](assets/favicon.svg)) is an independently redrawn optical master designed specifically for **32×32** and **16×16** pixel grids:

```
Favicon Grid: 32 × 32 px
Optical Stroke Unit: S = 4 px (12.5% stroke weight)
```

### Optical Adjustments vs. Master Mark
- **Bar Dimensions:** $28 \times 4\text{ px}$ ($7S \times S$), spanning $x \in [2, 30]$ and $y \in [2, 6]$, with pill end radius $r = 2\text{ px}$.
- **Hole Geometry:** Side $3S = 12\text{ px}$, top corners situated on $x = 10$ and $x = 22$, corner radii $r = 1\text{ px}$.
- **Saffron Drop:** Diameter $\varnothing\ 5\text{ px}$ ($1.25S$), positioned at $y \in [25, 30]$ with center at $(16, 27.5)$.
- **Gap to Tip:** $2.61\text{ px}$ ($0.65S$), enlarged to prevent pixel bridging.
- **Stroke-to-Width Ratio:** $0.143$ on favicon vs. $0.111$ on primary (**28.6% heavier stroke weight**).
- **Subpixel Guarantee:** Every corner rounding is $\ge 0.5\text{ px}$ at 16×16 resolution.
- **Aperture Clarity:** Yields **12 fully clear pixels** inside the triangular hole at 16×16 px (compared to only 4 clear pixels if the primary mark were naively scaled down).

---

## 4. Official Color Palette & Pairing Rules

```
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│                 │  │                 │  │                 │  │                 │
│   Ink Indigo    │  │  Saffron Truth  │  │   Warm Paper    │  │  Carbon Black   │
│     #12203F     │  │     #F2A33A     │  │     #F7F4EC     │  │     #000000     │
│                 │  │                 │  │                 │  │                 │
└─────────────────┘  └─────────────────┘  └─────────────────┘  └─────────────────┘
```

| Token Name | Hex Code | HSL | RGB | CMYK | Usage |
|---|---|---|---|---|---|
| `--vck-ink-indigo` | `#12203F` | `221°, 55%, 16%` | `18, 32, 63` | `71, 49, 0, 75` | Primary bar and V-funnel on white / light backgrounds |
| `--vck-saffron` | `#F2A33A` | `34°, 88%, 59%` | `242, 163, 58` | `0, 33, 76, 5` | Sifted truth drop across both light and dark variants |
| `--vck-warm-paper` | `#F7F4EC` | `45°, 41%, 95%` | `247, 244, 236` | `0, 1, 4, 3` | Reversed bar and V-funnel for dark backgrounds |
| `--vck-mono-carbon`| `#000000` | `0°, 0%, 0%` | `0, 0, 0` | `0, 0, 0, 100` | Single-color print, monochrome PDF rendering, terminal art |

### Pairing Rules & Contrast Governance

| Asset | Background Surface | Support Status | Visual Behavior |
|---|---|:---:|---|
| **`logo.svg`** (Primary) | White / Light (`#FFFFFF`, `#F8FAFC`) | **SUPPORTED (Canonical)** | Ink Indigo funnel with vibrant Saffron drop. |
| **`logo.svg`** (Primary) | Dark Navy (`#12203F`) | **UNSUPPORTED** | Bar and V vanish completely into the background. |
| **`logo-dark.svg`** (Reversed)| Dark Slate / Navy (`#0D1117`, `#12203F`) | **SUPPORTED (Canonical)** | Warm Paper funnel with vibrant Saffron drop. |
| **`logo-dark.svg`** (Reversed)| White / Light (`#FFFFFF`) | **UNSUPPORTED** | Warm paper on white has insufficient contrast. |
| **`logo-mono.svg`** (Mono) | White / Light (`#FFFFFF`) | **SUPPORTED** | Solid black funnel and solid black drop. |
| **`logo-mono.svg`** (Mono) | Dark (`#0D1117`, `#12203F`) | **UNSUPPORTED** | Black on dark has zero legibility. |
| **`favicon.svg`** (Optical) | Browser tabs, IDE status bars (16–32px)| **SUPPORTED** | High-contrast optical cut ensuring clear aperture. |

---

## 5. File Inventory & Implementation Guide

All production vector assets reside in [`docs/assets/`](assets/):

```
docs/assets/
├── logo.svg           ← Primary Mark (Ink Indigo + Saffron) for light backgrounds
├── logo-dark.svg      ← Reversed Mark (Warm Paper + Saffron) for dark backgrounds
├── logo-mono.svg      ← Monochrome Mark (Carbon Black) for single-color reproduction
└── favicon.svg        ← 32×32 Optical Cut for favicons, browser tabs, and desktop shortcuts
```

### GitHub Dual-Theme Integration
GitHub renders markdown using the viewer's active theme. Use the HTML5 `<picture>` element with `prefers-color-scheme`:

```html
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo-dark.svg">
    <img src="docs/assets/logo.svg" alt="Vivechak Logo" width="180" />
  </picture>
</p>
```

### Web & HTML Head Integration
```html
<!-- Scalable optical favicon for browser tabs -->
<link rel="icon" type="image/svg+xml" href="/assets/favicon.svg" />

<!-- Theme-adaptive icon selection -->
<link rel="icon" href="/assets/logo.svg" media="(prefers-color-scheme: light)" />
<link rel="icon" href="/assets/logo-dark.svg" media="(prefers-color-scheme: dark)" />
```

---

## 6. Clearspace & Brand Governance Rules

### Clearspace
A mandatory clearspace of **$1S$** ($36\text{ px}$ on master, $4\text{ px}$ on favicon) must surround the mark on all sides. No typography, grid lines, or layout borders may intrude into this boundary:

```
        ▲
        │  1S Clearspace (36 px)
        ▼
   ┌─────────────────────────────────────────┐
   │                                         │
◄──┤            [ VIVECHAK MARK ]            ├──► 1S Clearspace (36 px)
1S │                                         │
   └─────────────────────────────────────────┘
        ▲
        │  1S Clearspace (36 px)
        ▼
```

### Invariants
1. **Never omit the saffron dot:** The dot represents the core thesis of Vivechak: the surviving verified truth. Without it, the mark is incomplete.
2. **Never distort the aspect ratio:** The funnel arms are mathematically fixed at $30^\circ$ off vertical. Never squeeze, stretch, or skew.
3. **Never rotate:** The shirorekha headline bar must always sit horizontal and parallel to the layout plane.
4. **Never apply gradients:** Vivechak is built on empirical rigor. Fills must remain 100% solid flat colors.
5. **Respect contrast pairings:** Never place `logo.svg` on dark backgrounds or `logo-dark.svg` on light backgrounds.

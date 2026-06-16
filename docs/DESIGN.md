# YeboBank — Design System

This is the complete design specification. Every color, font, shadow, and spacing
decision is documented here with the reason. Follow it exactly.

## Color Palette

:root {
  /* === BACKGROUNDS === */
  --bg:           #f0ede4;   /* Page background. Warm cream. NOT white. Paper feeling. */
  --surface:      #ffffff;   /* Cards, modals, panels */
  --surface-2:    #f5f2e8;   /* Secondary panels, input backgrounds in dark contexts */
  --bg-dark:      #1a3a2a;   /* Dark sections (Bitcoin Join, Dark Stats, Footer) */

  /* === BRAND === */
  --forest:       #1a3a2a;   /* Primary. Nav, primary buttons, headings */
  --forest-mid:   #2d6b4a;   /* Hover state on forest elements */
  --forest-dim:   rgba(26,58,42,0.08); /* Tinted backgrounds on active nav items */
  --gold:         #b8a832;   /* Bitcoin/sats values SPECIFICALLY. Gold = money. */
  --gold-bright:  #c8d96a;   /* CTAs on dark (forest) backgrounds. Lime-gold. */
  --terra:        #c4622d;   /* Agent section, admin badge, secondary accent */

  /* === STATUS === */
  --green:        #2a7a4a;   /* Positive amounts, success states */
  --green-dim:    rgba(42,122,74,0.10);
  --red:          #c0392b;   /* Negative amounts, errors */
  --red-dim:      rgba(192,57,43,0.10);
  --amber:        #d97706;   /* Pending states, warnings */
  --amber-dim:    rgba(217,119,6,0.10);

  /* === TEXT === */
  --text:         #1a3a2a;   /* Primary text. Same hue as forest for harmony. */
  --text-muted:   #4a6a5a;   /* Secondary text, descriptions */
  --text-dim:     #8aaa9a;   /* Placeholder text, disabled, section labels */

  /* === BORDERS === */
  --border:       #e2ddd4;   /* Standard border */
  --border-light: #ece8e0;   /* Lighter dividers within cards */

  /* === GEOMETRY === */
  --radius-sm:    9px;       /* Inputs, badges, small buttons */
  --radius:       14px;      /* Cards, panels */
  --radius-lg:    20px;      /* Large cards, modals */
  --radius-xl:    28px;      /* Hero elements, balance card */

  /* === SHADOWS (THE VISUAL IDENTITY) === */
  /* Offset block shadow. NOT blurred. This is the design signature. */
  /* This renders cheaply on low-end Android phones (no GPU blur needed). */
  --shadow-sm:    4px  6px 0 0 rgba(26,58,42,0.08);
  --shadow:       8px 12px 0 0 rgba(26,58,42,0.10);
  --shadow-lg:   16px 24px 0 0 rgba(26,58,42,0.15);
}

## Typography

Two fonts only. Not three. Not one.

Space Grotesk (https://fonts.google.com/specimen/Space+Grotesk):
  Used for: ALL text. Body, headings, navigation, labels, buttons.
  Weights used: 400, 500, 600, 700, 800
  Load: @import from Google Fonts

JetBrains Mono (https://fonts.google.com/specimen/JetBrains+Mono):
  Used for: ALL numbers and codes. Nothing else.
  Weights used: 400, 500
  Load: @import from Google Fonts

What goes in JetBrains Mono:
- Satoshi amounts: 1,234,567 sats
- KES amounts: KES 4,320.50
- BTC amounts: 0.00041 BTC
- Sats notation: 41,000 sats
- Balance display on dashboard
- Lightning payment hashes (first 16 chars)
- Lightning invoice strings
- Transaction IDs in admin view
- BTC/KES rate display

What stays in Space Grotesk:
- Everything else

Font sizes:
  Hero headline: 40px / weight 800
  Section heading: 28px / weight 800
  Card heading: 18px / weight 700
  Body: 14px / weight 400
  Small/muted: 12px / weight 400
  Labels (uppercase): 10px / weight 700 / letter-spacing 0.1em

## The Balance Toggle

Every balance display shows three views. User taps to switch. No page reload.

View 1 (default): KES 4,320.50
View 2: 41,000 sats
View 3: 0.00041 BTC

Implementation: JavaScript on the page, reads data-sats attribute,
converts using current rate from data-kes-rate and data-sats-rate on body element.
Rate is injected server-side into the HTML.

HTML pattern:
  <body data-kes-rate="0.105" data-sats-rate="9.52">
  <div class="balance-amount" data-sats="41000">KES 4,320.50</div>
  <div class="balance-toggle">
    <button class="btog active" data-view="kes">KES</button>
    <button class="btog" data-view="sats">Sats</button>
    <button class="btog" data-view="btc">BTC</button>
  </div>

## Shadows: When to Use Each

--shadow-sm: cards in lists, table-like card rows
--shadow: primary cards, form panels, phone mockup
--shadow-lg: hero phone mockup, floating cards, CTA banners

Never use box-shadow with blur radius. Always use the offset style.
Never apply shadows to: navigation bars, table rows, form inputs, text.

## Navigation

### Desktop Sidebar (visible at 1024px+)
Maximum 6 items per section. Two sections maximum per role.

Customer sidebar:
  WALLET
    Dashboard
    Add money (deposit)
    Send / Receive (combined page with two tabs)
    History
  SAVINGS
    My savings
  ACCOUNT
    Settings

Agent sidebar:
  AGENT
    Dashboard
    Accept cash (cash-in)
    Pay out cash (cash-out)
  WALLET
    My wallet

### Mobile Bottom Navigation (visible below 1024px)
Maximum 5 items. Required for Kenya — most users are on phones.

Customer bottom nav:
  Home (dashboard icon)
  Send/Receive (send icon)
  Savings (lock icon)
  Agent (if user is agent: person icon)
  Account (settings icon)

## Icon System

Use Tabler Icons: https://tabler.io/icons
Load via CDN: https://cdn.jsdelivr.net/npm/@tabler/icons-webfont@latest/tabler-icons.min.css
Use as: <i class="ti ti-arrow-down"></i>

Icons required:
  ti-arrow-down         Deposit / credit
  ti-arrow-up           Withdrawal / debit
  ti-send               Send money
  ti-qrcode             Receive / scan
  ti-list               Transaction history
  ti-lock               Savings lock
  ti-star               Interest
  ti-users              Chama group
  ti-user               Individual member
  ti-building-bank      Admin / institutional
  ti-trending-up        Growth / yield
  ti-currency-bitcoin   BTC indicator (use sparingly)
  ti-shield-check       Security / verified
  ti-settings           Settings
  ti-logout             Sign out
  ti-check              Success / confirm
  ti-x                  Error / cancel
  ti-clock              Pending
  ti-chart-pie          Treasury breakdown
  ti-cash               Agent cash
  ti-map-pin            Agent location
  ti-sparkles           Interest / reward (for monthly interest credit)

## Responsive Breakpoints

Mobile first. Add complexity as screen grows.

Default (0px+): mobile layout, bottom nav, single column
sm (640px+): larger phone, minor padding adjustments
md (768px+): tablet, two-column grids appear
lg (1024px+): desktop, sidebar appears, bottom nav hidden

## Component Patterns

### Balance Card (dark, on forest background)
- Background: var(--forest)
- Text: white
- Accent glow: radial-gradient terra-colored in corner
- Balance display: JetBrains Mono, 2.4rem
- Shadow: var(--shadow-lg) on the card itself
- Shows: KES / Sats / BTC toggle, quick action buttons below

### Quick Action Buttons (below balance card)
- Background: rgba(255,255,255,0.1)
- Text: rgba(255,255,255,0.85)
- Hover: rgba(255,255,255,0.18)
- Layout: flex row, gap: 8px
- Primary action: --terra background

### Transaction Row
- No card shadow on individual rows
- Divider: 1px var(--border-light) between rows
- Positive amounts: var(--green)
- Negative amounts: var(--red)
- Pending amounts: var(--amber)
- Time: var(--text-dim), 12px

### Status Badges
- border-radius: 20px (fully rounded)
- padding: 3px 9px
- font-size: 10px
- font-weight: 700
- text-transform: uppercase
- letter-spacing: 0.04em

Badge styles:
  confirmed / credit / active: green-dim background, green text
  pending: amber-dim background, amber text
  failed / debit: red-dim background, red text
  agent: terra-dim background, terra text
  admin: gold-dim background, gold text

## What NOT to Build

Do NOT add:
- Dark mode (cream + forest is the brand — dark mode breaks the visual identity)
- Animated transitions longer than 0.2s (performance on low-end Android)
- Custom fonts beyond Space Grotesk and JetBrains Mono
- More than two accent colors on any single screen
- Box shadows with blur on mobile (expensive to render)
- Gradients except on the balance card corner accent
- Images or illustrations that are not community/Kenya-specific

## Accessibility Requirements

- All interactive elements: minimum 44x44px touch target
- Color is never the only indicator of status (always pair with text or icon)
- Inputs: visible focus ring (3px forest-colored glow)
- Buttons: descriptive aria-label when icon-only
- Images: alt text required

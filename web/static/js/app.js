// YeboBank — app.js
// KES↔Sats converter, balance display toggle (Sats / KES / BTC)
// Zero dependencies. Vanilla JS only.

(function () {
  'use strict';

  // ── Balance toggle ─────────────────────────────────────────────────────────
  const MODES = ['sats', 'kes', 'btc'];
  let modeIdx = 0;

  function initBalanceToggle() {
    const toggle = document.querySelector('.balance-toggle');
    const satsEl = document.querySelector('.balance-sats');
    const kesEl  = document.querySelector('.balance-kes');
    if (!toggle || !satsEl) return;

    const sats   = parseInt(satsEl.dataset.sats  || '0', 10);
    const btcKES = parseFloat(document.body.dataset.btcKes || '0');
    const kes    = sats / 1e8 * btcKES;
    const btc    = sats / 1e8;

    function update() {
      const mode = MODES[modeIdx];
      if (mode === 'sats') {
        satsEl.textContent = formatSats(sats) + ' sats';
        if (kesEl) kesEl.textContent = '≈ KES ' + formatKES(kes);
        toggle.textContent = 'show KES →';
      } else if (mode === 'kes') {
        satsEl.textContent = 'KES ' + formatKES(kes);
        if (kesEl) kesEl.textContent = '≈ ' + formatSats(sats) + ' sats';
        toggle.textContent = 'show BTC →';
      } else {
        satsEl.textContent = '₿ ' + btc.toFixed(8);
        if (kesEl) kesEl.textContent = '≈ KES ' + formatKES(kes);
        toggle.textContent = 'show sats →';
      }
    }

    toggle.addEventListener('click', function () {
      modeIdx = (modeIdx + 1) % MODES.length;
      update();
    });
    update();
  }

  // ── KES↔Sats live conversion preview ──────────────────────────────────────
  function initConverter() {
    const btcKES = parseFloat(document.body.dataset.btcKes || '0');
    if (!btcKES) return;

    document.querySelectorAll('[data-convert-from="kes"]').forEach(function (input) {
      const targetId = input.dataset.convertTarget;
      const target = targetId ? document.getElementById(targetId) : null;
      if (!target) return;
      input.addEventListener('input', function () {
        const kes = parseFloat(this.value) || 0;
        const sats = Math.round(kes / btcKES * 1e8);
        target.textContent = formatSats(sats) + ' sats';
      });
    });

    document.querySelectorAll('[data-convert-from="sats"]').forEach(function (input) {
      const targetId = input.dataset.convertTarget;
      const target = targetId ? document.getElementById(targetId) : null;
      if (!target) return;
      input.addEventListener('input', function () {
        const sats = parseInt(this.value, 10) || 0;
        const kes = sats / 1e8 * btcKES;
        target.textContent = '≈ KES ' + formatKES(kes);
      });
    });
  }

  // ── Formatters ─────────────────────────────────────────────────────────────
  function formatSats(n) {
    return n.toLocaleString('en-KE');
  }

  function formatKES(n) {
    return n.toLocaleString('en-KE', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  // ── Flash message auto-dismiss ─────────────────────────────────────────────
  function initAlerts() {
    document.querySelectorAll('.alert').forEach(function (el) {
      setTimeout(function () {
        el.style.transition = 'opacity 0.4s';
        el.style.opacity = '0';
        setTimeout(function () { el.remove(); }, 400);
      }, 5000);
    });
  }

  // ── Init ───────────────────────────────────────────────────────────────────
  document.addEventListener('DOMContentLoaded', function () {
    initBalanceToggle();
    initConverter();
    initAlerts();
  });
}());

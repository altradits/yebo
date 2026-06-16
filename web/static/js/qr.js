// YeboBank — qr.js
// QR code generation using qrcode.js loaded from CDN.
// Usage: <div id="qr" data-value="lnbc..."></div>
// Include qrcode.js CDN before this file.

(function () {
  'use strict';

  document.addEventListener('DOMContentLoaded', function () {
    var el = document.getElementById('qr');
    if (!el || typeof QRCode === 'undefined') return;

    var value = el.dataset.value;
    if (!value) return;

    new QRCode(el, {
      text: value,
      width: 240,
      height: 240,
      colorDark: '#1a3a2a',
      colorLight: '#ffffff',
      correctLevel: QRCode.CorrectLevel.M
    });
  });
}());

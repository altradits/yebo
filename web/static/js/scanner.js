// YeboBank — scanner.js
// Camera-based QR scanner for Lightning invoices.
// Uses the browser's getUserMedia API (no library needed for modern browsers).

(function () {
  'use strict';

  var scanBtn    = document.getElementById('scan-btn');
  var videoEl    = document.getElementById('scan-video');
  var resultEl   = document.getElementById('scan-result');
  var inputEl    = document.getElementById('payment_request');

  if (!scanBtn || !videoEl) return;

  scanBtn.addEventListener('click', function () {
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      alert('Camera not supported on this device.');
      return;
    }
    videoEl.style.display = 'block';
    navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } })
      .then(function (stream) {
        videoEl.srcObject = stream;
        videoEl.play();
        scanFrames(stream);
      })
      .catch(function (err) {
        alert('Camera error: ' + err.message);
      });
  });

  function scanFrames(stream) {
    if (!window.BarcodeDetector) {
      // Fallback: user pastes manually
      stopStream(stream);
      if (resultEl) resultEl.textContent = 'QR scan not supported. Please paste the invoice.';
      return;
    }
    var detector = new BarcodeDetector({ formats: ['qr_code'] });
    var canvas   = document.createElement('canvas');
    var ctx      = canvas.getContext('2d');

    function tick() {
      if (videoEl.readyState !== videoEl.HAVE_ENOUGH_DATA) {
        requestAnimationFrame(tick);
        return;
      }
      canvas.width  = videoEl.videoWidth;
      canvas.height = videoEl.videoHeight;
      ctx.drawImage(videoEl, 0, 0);
      detector.detect(canvas).then(function (barcodes) {
        if (barcodes.length > 0) {
          var value = barcodes[0].rawValue;
          if (inputEl) inputEl.value = value;
          if (resultEl) resultEl.textContent = 'Scanned: ' + value.slice(0, 40) + '...';
          stopStream(stream);
        } else {
          requestAnimationFrame(tick);
        }
      }).catch(function () {
        requestAnimationFrame(tick);
      });
    }
    requestAnimationFrame(tick);
  }

  function stopStream(stream) {
    stream.getTracks().forEach(function (t) { t.stop(); });
    videoEl.style.display = 'none';
  }
}());

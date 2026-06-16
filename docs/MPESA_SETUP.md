# M-Pesa (Daraja) Setup

## Sandbox Registration

1. Go to https://developer.safaricom.co.ke
2. Create an account and log in
3. Go to "My Apps" → "Add New App"
4. Enable: **Mpesa Sandbox** and **Lipa Na Mpesa Sandbox**
5. Copy your `Consumer Key` and `Consumer Secret`
6. Set in `.env`:
   ```
   MPESA_ENV=sandbox
   MPESA_CONSUMER_KEY=your_consumer_key
   MPESA_CONSUMER_SECRET=your_consumer_secret
   MPESA_SHORTCODE=174379
   MPESA_PASSKEY=bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919
   MPESA_CALLBACK_URL=https://your-ngrok-url.ngrok.io
   ```

## Sandbox Test Credentials

- Test phone: `254708374149`
- Test PIN: any 4-digit number
- STK Push will simulate on this number

## Local Callback Testing

Use ngrok to expose your local server:
```bash
ngrok http 8080
# Copy the https URL → set as MPESA_CALLBACK_URL
```

## Production Go-Live

1. Apply for go-live at https://developer.safaricom.co.ke/APIs/MpesaExpressSimulate
2. Submit business registration documents
3. Safaricom reviews within 5-10 business days
4. Update `.env`:
   ```
   MPESA_ENV=production
   MPESA_SHORTCODE=your_production_shortcode
   MPESA_PASSKEY=your_production_passkey
   ```
5. Configure nginx to allowlist Safaricom IPs:
   ```nginx
   allow 196.201.214.0/24;
   deny all;
   ```

## STK Push Flow

1. User enters amount + phone in `/deposit`
2. Server calls Daraja `/mpesa/stkpush/v1/processrequest`
3. User receives PIN prompt on phone
4. User enters PIN
5. Safaricom POSTs result to `/webhook/mpesa`
6. Webhook validates, checks idempotency, credits wallet

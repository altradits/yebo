# YeboBank — M-Pesa Daraja Integration

## Overview

M-Pesa is the on-ramp and off-ramp for KES. Users send KES via M-Pesa, we convert to sats
at the current BTC/KES rate and credit their wallet. Withdrawals reverse the flow.

## Setup (Do This First)

Step 1: Register at https://developer.safaricom.co.ke (free, instant)
Step 2: Create an app in the dashboard
Step 3: Get sandbox credentials:
  - Consumer Key
  - Consumer Secret
  - Test Shortcode: 174379
  - Test Passkey: bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919
  - Test Phone (for STK Push): 254708374149

Step 4: For production access, you need:
  - Registered Kenyan business + KRA PIN
  - Apply through a Safaricom M-Pesa API partner
  - CBK sandbox authorization for financial services

## Environment Variables

MPESA_ENV=sandbox              # 'sandbox' or 'production'
MPESA_CONSUMER_KEY=            # from Daraja dashboard
MPESA_CONSUMER_SECRET=         # from Daraja dashboard
MPESA_SHORTCODE=174379         # sandbox shortcode
MPESA_PASSKEY=bfb279f9...      # sandbox passkey
MPESA_CALLBACK_URL=https://yebobank.com  # MUST be HTTPS, reachable from internet

## Sandbox vs Production URLs

Sandbox:    https://sandbox.safaricom.co.ke
Production: https://api.safaricom.co.ke

## API Endpoints Used

### OAuth Token
GET /oauth/v1/generate?grant_type=client_credentials
Authorization: Basic base64(consumer_key:consumer_secret)
Response: { "access_token": "...", "expires_in": "3599" }
Cache this token for 58 minutes, then refresh.

### STK Push (Customer Deposit)
POST /mpesa/stkpush/v1/processrequest
This sends a payment prompt to the customer's phone.
Customer sees: "YeboBank requests KES X from your M-Pesa"
Customer enters PIN on their phone.
Safaricom calls your callback URL with result.

Request body:
{
  "BusinessShortCode": "174379",
  "Password": base64(shortcode + passkey + timestamp),
  "Timestamp": "20240115123045",
  "TransactionType": "CustomerPayBillOnline",
  "Amount": "500",
  "PartyA": "254712345678",
  "PartyB": "174379",
  "PhoneNumber": "254712345678",
  "CallBackURL": "https://yebobank.com/api/mpesa/callback",
  "AccountReference": "user-account-id",
  "TransactionDesc": "YeboBank deposit"
}

Response on success:
{
  "MerchantRequestID": "...",
  "CheckoutRequestID": "ws_CO_...",
  "ResponseCode": "0",
  "ResponseDescription": "Success. Request accepted for processing",
  "CustomerMessage": "Success. Request accepted for processing"
}

### B2C (Withdrawal to Customer)
POST /mpesa/b2c/v3/paymentrequest
This sends money from your shortcode to a customer's M-Pesa.
Requires B2C API access (production requires higher-tier access).

## Callback Handler

Safaricom sends POST to your CallbackURL when a transaction completes.
Your handler MUST:
1. Respond HTTP 200 immediately (Safaricom retries on non-200)
2. Parse the callback JSON
3. Check idempotency: SELECT id FROM mpesa_transactions WHERE mpesa_reference = $receipt
4. If receipt already processed: log + return 200
5. Find the pending mpesa_transaction by CheckoutRequestID
6. Get current BTC/KES rate
7. Calculate sats = KES * sats_per_kes
8. db.Begin()
9. db.CreditSats(tx, walletID, sats, "deposit_mpesa", ...)
10. UPDATE mpesa_transactions SET status='confirmed', mpesa_reference=$receipt ...
11. db.Commit()
12. Return 200

Callback JSON structure (success):
{
  "Body": {
    "stkCallback": {
      "MerchantRequestID": "...",
      "CheckoutRequestID": "ws_CO_...",
      "ResultCode": 0,
      "ResultDesc": "The service request is processed successfully.",
      "CallbackMetadata": {
        "Item": [
          {"Name": "Amount", "Value": 500},
          {"Name": "MpesaReceiptNumber", "Value": "QKL8XNHJP4"},
          {"Name": "TransactionDate", "Value": 20240115123456},
          {"Name": "PhoneNumber", "Value": 254712345678}
        ]
      }
    }
  }
}

Callback JSON structure (failure/cancelled):
{
  "Body": {
    "stkCallback": {
      "MerchantRequestID": "...",
      "CheckoutRequestID": "ws_CO_...",
      "ResultCode": 1032,
      "ResultDesc": "Request cancelled by user."
    }
  }
}
ResultCode 0 = success. Any other = failure.

## Phone Number Format

Always use format: 2547XXXXXXXX (12 digits, no +, no 07)
Convert: 0712345678 → 254712345678
Convert: +254712345678 → 254712345678

## Testing in Sandbox

Use Safaricom test credentials. STK Push in sandbox does NOT send a real prompt.
You must simulate the callback by POST-ing to your own callback URL.

Test callback simulator:
POST https://sandbox.safaricom.co.ke/mpesa/stkpush/v1/checkoutrequestquery
to check status of a sandbox STK Push.

## Common Errors

ResultCode 1032: User cancelled the payment prompt
ResultCode 1037: DS timeout / user did not respond
ResultCode 2001: Wrong PIN entered
ResultCode 1: Insufficient funds in customer M-Pesa
HTTP 400 from Daraja: Usually invalid timestamp format or wrong passkey

# YeboBank API Reference

All HTML routes. No REST JSON API for external consumers — this is a server-rendered web app.
The only JSON endpoints are webhooks and LNURL-pay (called by Lightning wallets).

## Public Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Marketing landing page |
| GET | `/login` | Login form |
| POST | `/login` | Authenticate (phone + password) |
| GET | `/register` | Registration form |
| POST | `/register` | Create account |
| GET | `/logout` | Destroy session, redirect to /login |
| GET | `/.well-known/lnurlp/{username}` | LNURL-pay step 1 (JSON) |
| GET | `/lnurl/pay/{username}/callback` | LNURL-pay step 2 — returns invoice (JSON) |
| POST | `/webhook/mpesa` | Safaricom STK Push callback (JSON) |
| POST | `/webhook/lnd` | LND invoice settled webhook |

## Customer Routes (authenticated)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/dashboard` | Main dashboard with balance |
| GET | `/deposit` | Deposit options |
| POST | `/deposit/mpesa` | Initiate M-Pesa STK Push |
| POST | `/deposit/lightning` | Generate Lightning invoice |
| GET | `/withdraw` | Withdrawal options |
| POST | `/withdraw/mpesa` | B2C payout to phone |
| POST | `/withdraw/lightning` | Pay Lightning invoice |
| GET | `/history` | Transaction history |
| GET | `/savings` | Savings overview + locks |
| GET | `/savings/lock` | Lock form |
| POST | `/savings/lock` | Create savings lock |
| GET | `/chama` | Group savings list |
| GET | `/chama/create` | Create chama form |
| POST | `/chama/create` | Create new chama |
| POST | `/chama/contribute` | Contribute to a chama |
| GET | `/settings` | Account settings |
| POST | `/settings` | Update account settings |

## Agent Routes (role: agent)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/agent` | Agent dashboard |
| GET | `/agent/cashin` | Cash-in form |
| POST | `/agent/cashin` | Credit customer, record commission |
| GET | `/agent/cashout` | Cash-out form |
| POST | `/agent/cashout` | Debit customer |

## Trader Routes (role: trader)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/trader` | Treasury overview |
| GET | `/trader/assets` | Treasury asset list |
| POST | `/trader/distribute` | Manually trigger interest distribution |
| GET | `/trader/profit` | Distribution history |

## Admin Routes (role: admin)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/admin` | Admin dashboard |
| GET | `/admin/customers` | Customer list |
| POST | `/admin/customers/toggle` | Activate / suspend user |
| GET | `/admin/agents` | Agent list |
| POST | `/admin/agents/approve` | Approve pending agent |
| GET | `/admin/settings` | Pool settings |
| POST | `/admin/settings` | Update pool settings |

## M-Pesa Callback Format (from Safaricom)

```json
{
  "Body": {
    "stkCallback": {
      "MerchantRequestID": "...",
      "CheckoutRequestID": "...",
      "ResultCode": 0,
      "ResultDesc": "The service request is processed successfully.",
      "CallbackMetadata": {
        "Item": [
          { "Name": "Amount",              "Value": 500 },
          { "Name": "MpesaReceiptNumber",  "Value": "NLJ7RT61SV" },
          { "Name": "PhoneNumber",         "Value": 254708374149 }
        ]
      }
    }
  }
}
```

## Error Codes

All errors are displayed inline in the HTML form. No custom error codes — HTTP status codes only:
- `200` — success
- `302` — redirect (after form POST)
- `400` — bad request
- `401` — unauthenticated (redirects to /login)
- `403` — forbidden (wrong role)
- `429` — rate limited
- `500` — server error

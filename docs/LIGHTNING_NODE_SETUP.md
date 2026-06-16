# Lightning Node Setup

YeboBank requires an LND node with REST API access. Two options:

## Option A — Voltage.cloud (Recommended for getting started)

1. Sign up at https://voltage.cloud
2. Create an LND node (Standard plan, ~$12/month)
3. Download `admin.macaroon` and note your node URL
4. Set in `.env`:
   ```
   LND_REST_URL=https://your-node.voltageapp.io:8080
   LND_MACAROON_PATH=/path/to/admin.macaroon
   ```
5. No `LND_TLS_CERT_PATH` needed — Voltage uses valid TLS.

## Option B — Self-hosted LND

1. Install LND: https://github.com/lightningnetwork/lnd/releases
2. Configure `lnd.conf`:
   ```
   [Application Options]
   restlisten=0.0.0.0:8080
   [Bitcoin]
   bitcoin.active=1
   bitcoin.mainnet=1
   bitcoin.node=neutrino
   [Neutrino]
   neutrino.connect=btcd-mainnet.lightning.computer
   ```
3. Start LND: `lnd`
4. Create wallet: `lncli create`
5. Enable watchtower (required by whitepaper §1):
   ```
   [Wtclient]
   wtclient.active=1
   ```
6. Set in `.env`:
   ```
   LND_REST_URL=https://localhost:8080
   LND_MACAROON_PATH=/home/ubuntu/.lnd/data/chain/bitcoin/mainnet/admin.macaroon
   LND_TLS_CERT_PATH=/home/ubuntu/.lnd/tls.cert
   ```

## Channel Management

- Open channels with well-connected peers (ACINQ, Bitrefill, WalletOfSatoshi)
- Minimum channel size: 500,000 sats
- Keep inbound liquidity equal to expected monthly deposit volume
- Monitor channel health: `lncli listchannels`

## Watchtower (Required)

```
lncli wtclient add <watchtower-pubkey>@<ip>:<port>
```

Public watchtowers: https://github.com/lightningnetwork/lnd/blob/master/watchtower/wtclient/README.md

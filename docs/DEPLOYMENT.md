# Deployment Guide

## Docker Compose (Recommended)

```bash
git clone https://github.com/yebobank/yebobank
cd yebobank
cp .env.example .env
# Edit .env with your credentials

docker compose up -d
```

Access at `http://localhost:8080`.

## nginx + TLS (Production)

Create `/etc/nginx/sites-available/yebobank`:

```nginx
server {
    listen 80;
    server_name yebobank.com;
    location /.well-known/acme-challenge/ { root /var/www/certbot; }
    location / { return 301 https://$host$request_uri; }
}

server {
    listen 443 ssl;
    server_name yebobank.com;

    ssl_certificate     /etc/letsencrypt/live/yebobank.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yebobank.com/privkey.pem;

    # M-Pesa callback endpoint — Safaricom IPs only
    location /webhook/mpesa {
        allow 196.201.214.0/24;
        allow 196.201.214.136/29;
        deny all;
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header Host $host;
    }
}
```

Get TLS certificate:
```bash
certbot --nginx -d yebobank.com
```

## Bare Metal (systemd)

Create `/etc/systemd/system/yebobank.service`:

```ini
[Unit]
Description=YeboBank
After=network.target postgresql.service

[Service]
Type=simple
User=yebo
WorkingDirectory=/opt/yebobank
EnvironmentFile=/opt/yebobank/.env
ExecStart=/opt/yebobank/yebobank
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
systemctl enable --now yebobank
```

## Environment Variables

See `.env.example` for the complete list.

## Production Checklist

- [ ] `ENV=production` set
- [ ] `DB_URL` points to production PostgreSQL
- [ ] LND node funded and channels open
- [ ] M-Pesa go-live approved by Safaricom
- [ ] `MPESA_CALLBACK_URL` is the HTTPS production URL
- [ ] nginx configured with Safaricom IP allowlist
- [ ] TLS certificate active and auto-renewing
- [ ] Backups configured for PostgreSQL
- [ ] LND watchtower active
- [ ] Admin account created (`ADMIN_PHONE` + `ADMIN_PASSWORD` in .env)

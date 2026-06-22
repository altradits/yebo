-- OTP requests: short-lived codes for phone-based login
CREATE TABLE otp_requests (
    phone       TEXT        PRIMARY KEY,
    otp         TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL
);

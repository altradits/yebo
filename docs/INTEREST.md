# YeboBank — Interest Engine

## Overview

Users lock sats for a minimum of 5 years. On the 2nd of each month, a background goroutine
runs the distribution. It calculates how much each lock earns proportionally from the pool
profit logged by the trader, deducts the 2% admin fee, and credits each wallet in a single
atomic database transaction.

## The Math

Monthly rate = (1 + annual_rate)^(1/12) - 1
At 5.2% APY: (1 + 0.052)^(1/12) - 1 = 0.004232 = approximately 0.423% per month

For a lock of 1,000,000 sats: 1,000,000 * 0.00423 = 4,230 sats per month

BUT: YeboBank does not auto-calculate yield from a formula.
The TRADER logs actual profit from real assets (bonds, money market, equities).
Distribution is based on ACTUAL pool profit, not a formula.
The 5.2% APY is a TARGET, not a guarantee.

Distribution formula:
- Each lock earns: (lock_principal / total_locked) * distributable_sats
- distributable_sats = pool_profit_sats - admin_fee_sats
- admin_fee_sats = pool_profit_sats * 0.02 (2%)
- Rounding remainder goes to admin (never lost)

Example:
  Pool profit this month: 500,000 sats
  Admin fee (2%): 10,000 sats
  Distributable: 490,000 sats
  
  User A lock: 5,000,000 sats
  User B lock: 3,000,000 sats
  User C lock: 2,000,000 sats
  Total locked: 10,000,000 sats
  
  User A earns: (5,000,000 / 10,000,000) * 490,000 = 245,000 sats
  User B earns: (3,000,000 / 10,000,000) * 490,000 = 147,000 sats
  User C earns: (2,000,000 / 10,000,000) * 490,000 = 98,000 sats
  Total: 490,000 sats (exact, no rounding in this example)

## Early Exit Penalties

If a user exits before maturity, they forfeit a percentage of their accrued interest.
Principal is ALWAYS returned in full. Penalty only applies to earned interest.

Years elapsed | Penalty on accrued interest
Under 1 year  | 100% (forfeit all interest earned so far)
1-2 years     | 75%
2-3 years     | 50%
3-4 years     | 25%
4-5 years     | 10%
5+ years      | 0% (matured — no penalty)

Calculation:
  penalty_sats = accrued_sats * penalty_pct / 100
  returned_sats = principal_sats + accrued_sats - penalty_sats
  forfeited_sats = penalty_sats -> goes to admin wallet

## Scheduler

Runs on: 2nd of each month at 00:01 UTC
(2nd not 1st, to ensure all profit_logs for the previous month are submitted)

If distribution fails: retries every hour for up to 24 hours.
If all retries fail: logs error, alerts admin. Manual trigger available in admin dashboard.

The goroutine is started in cmd/server/main.go at startup.
There is one goroutine per server instance. 
If running multiple instances, only ONE should run the scheduler (use a DB lock):

  -- In postgres, use advisory lock to ensure only one instance runs distribution
  SELECT pg_try_advisory_lock(12345) -- 12345 is an arbitrary unique ID for this task
  If returns false: another instance already has the lock, skip this run.

## Trader Workflow (Before Distribution Can Run)

Each month, the trader must:
1. Log into /trader/profit
2. Enter total profit earned from pool assets this month
3. Enter which assets generated the profit and amounts
4. Click "Record profit"
5. This inserts a row into profit_logs

The distribution engine reads profit_logs for the current month.
If no profit is logged: distribution returns an error and does not run.

## Test Cases (Write These First)

- Zero active locks: distribution should return error "no active locks"
- One lock only: that lock gets 100% of distributable (minus admin fee)
- 1000 locks equal principal: each gets equal share
- Rounding: total distributed + admin fee MUST equal pool_profit_sats exactly
- Distribution already run this month: second run returns error (UNIQUE on period_month)
- Lock added mid-month: included if locked_at < start of distribution month
- Early exit before distribution: lock excluded from distribution (status != 'active')

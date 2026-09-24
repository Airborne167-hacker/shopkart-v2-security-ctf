# Facilitator Answer Key (spoilers — don't share with players)

## Flag 1 — The Front Door (SQL Injection)
Go to `/login`. In the **Username** field, enter:
```
admin' --
```
Leave the password blank (or anything). This closes the SQL string early and
comments out the password check, logging you in as `admin`. Go to
`/account` to see the flag.

**Flag:** `WIN{a0ca3a0bb29c61f947c8}`

## Flag 2 — The Spare Key (Broken Authentication)
Go to `/about`, then View Page Source (Ctrl+U / Cmd+Option+U). There's a
hidden `<div>` (not visible on the rendered page) containing:
```
username: staff_alerts
password: ShopKart@123
```
Log in with those at `/login` normally (no injection needed). Go to
`/account` to see the flag.

**Flag:** `WIN{12b332de4e84e4583f20}`

## Flag 3 — The Poisoned Review (Stored XSS)
Go to any product page, e.g. `/product/1`. In the review form, submit a
review whose **content** field is exactly:
```
<script>alert(SHOPKART_SECRET)</script>
```
The page renders review content unescaped, so this actually executes and
pops an alert box containing the flag.

**Flag:** `WIN{163f7e4d8deb3cd999ed}`

## Flag 4 — The Leaky API (Sensitive Data Exposure)
Visit `/api/users` directly. It's a JSON dump of every user account,
including a `system_backup` account with an `internal_notes` field
containing the flag.

**Flag:** `WIN{a3902d54465a301968a8}`

## Flag 5 — The Backdoor Left Open (Security Misconfiguration)
Visit `/.env` directly. It's a plaintext "environment file" with fake DB
credentials and the flag on the last line.

**Flag:** `WIN{4fbecca978479827467f}`

---

All 5 flags can also be captured in any order, and hints (3 per challenge,
progressively more specific) are available in-app via the "Ask the Security
Consultant" button on each challenge card.

# LinkedIn Checkpoint Guide

## What's Happening?

LinkedIn is detecting automation and showing **checkpoint/challenge pages** that require manual verification. This is LinkedIn's security system protecting against bots.

## Why This Happens

1. **New account** - Fresh accounts trigger more checkpoints
2. **Automation detection** - LinkedIn's systems detect browser automation
3. **Unusual activity** - Rapid actions look suspicious
4. **IP/Device reputation** - New devices/IPs get more scrutiny

## What the Logs Show

```
level=info msg="post‑login page URL" url="...checkpoint/challenge..."
level=warning msg="detected checkpoint page"
level=warning msg="context canceled"
level=info msg="search completed" total_profiles=0
```

This means:
- ✅ Login worked
- ⚠️ LinkedIn showed a checkpoint
- ❌ Search was blocked
- ❌ No profiles found

## Solutions

### Immediate Fix (Manual)

1. **When you see checkpoint in browser:**
   - Complete the verification (phone, email, captcha, etc.)
   - Wait for it to redirect to your feed
   - Then the app can continue

2. **The app now waits 30 seconds** when it detects a checkpoint
   - Use this time to complete verification
   - The app will check again and continue if resolved

### Long-term Solutions

#### 1. Warm Up Your Account
- Log in manually to LinkedIn 5-10 times over a few days
- Browse profiles, send a few manual connection requests
- Let LinkedIn "trust" your account first
- Then run automation

#### 2. Use Conservative Settings
```yaml
# In config.yaml - make delays longer
search:
  page_delay_min: 10s  # Was 2s
  page_delay_max: 20s  # Was 5s

connect:
  daily_limit: 2  # Was 5 - start very small
  action_delay_min: 10s  # Was 3s
  action_delay_max: 20s  # Was 8s
```

#### 3. Run Less Frequently
- **Don't run multiple times per day**
- Wait 24-48 hours between runs
- LinkedIn has weekly limits (~100 connection requests/week)

#### 4. Use a Mature Account
- Accounts older than 6 months get fewer checkpoints
- Accounts with real activity (posts, connections) are more trusted
- Premium accounts may have fewer restrictions

## How to Use the App Now

### Step 1: Run the App
```cmd
set LINKEDIN_EMAIL=your-email@example.com
set LINKEDIN_PASSWORD=your-password
go run ./cmd/app
```

### Step 2: Watch the Browser
- Keep `headless: false` in config.yaml
- When checkpoint appears, **complete it manually**
- The app will wait 30 seconds for you

### Step 3: Let It Continue
- After checkpoint is resolved, the app continues automatically
- If still blocked, it will skip and log a warning

## Expected Behavior

### Good Run (No Checkpoint):
```
✅ Login successful
✅ Search found 15 profiles
✅ Sent 3 connection requests
✅ Completed successfully
```

### Checkpoint Run:
```
✅ Login successful
⚠️ Checkpoint detected - waiting 30 seconds
   [You complete verification here]
✅ Checkpoint resolved
✅ Search found profiles
✅ Sent connection requests
```

### Blocked Run:
```
✅ Login successful
⚠️ Checkpoint detected
❌ Still on checkpoint after wait
⚠️ Skipping search
⚠️ No profiles found
```

## Best Practices

1. **First Time:**
   - Run with browser visible (`headless: false`)
   - Complete any checkpoints manually
   - Let it finish one full run
   - Check results in database

2. **Regular Use:**
   - Run once per day maximum
   - Use very conservative limits (2-3 requests/day)
   - Wait if you see checkpoints
   - Don't force it

3. **If Checkpoints Persist:**
   - Stop automation for 48 hours
   - Use LinkedIn normally (manual browsing)
   - Then try again with even smaller limits

## Understanding the Limits

LinkedIn has **hard limits** you cannot bypass:
- **Connection requests:** ~100 per week
- **Messages:** ~50-100 per week (varies)
- **Search:** Unlimited, but checkpoints may appear

If you hit limits:
- Wait 7 days
- Reduce your daily limits
- Use LinkedIn manually to "reset" your reputation

## Troubleshooting

### "Always getting checkpoints"
→ Account is too new or suspicious
→ Solution: Warm up account manually for a week

### "Context canceled errors"
→ LinkedIn is blocking navigation
→ Solution: Complete checkpoints, wait longer between runs

### "0 profiles found"
→ Checkpoint blocking search OR page didn't load
→ Solution: Check browser window, complete verification

### "Cookies not saving"
→ This is a known limitation (DevTools protocol issue)
→ Solution: You'll need to log in each time (not a big deal for PoC)

## The Reality

This is a **proof-of-concept**, not production software. LinkedIn actively fights automation, so:

- ✅ **Expected:** Some checkpoints, occasional failures
- ✅ **Normal:** Manual intervention needed sometimes  
- ❌ **Not expected:** 100% automation without any checkpoints
- ❌ **Not realistic:** Running 24/7 without issues

## Success Criteria

A successful run means:
- ✅ App doesn't crash
- ✅ Handles errors gracefully
- ✅ Finds some profiles (even if 0 due to checkpoints)
- ✅ Sends requests when possible
- ✅ Logs everything clearly

**The app is working correctly** - LinkedIn's security is just doing its job.

## Next Steps

1. **Try running again** with the new checkpoint detection
2. **Complete checkpoints manually** when they appear
3. **Use very conservative settings** in config.yaml
4. **Be patient** - automation detection is a cat-and-mouse game

Good luck! 🚀


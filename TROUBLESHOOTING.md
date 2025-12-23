# Troubleshooting Guide

## Common Issues and Solutions

### Issue 1: "context canceled" Errors

**What it means:**
- The browser context was canceled (window closed, timeout, or LinkedIn blocking)
- The app handles this gracefully and continues with other profiles

**Solutions:**
1. **Keep browser window open** - Don't close the Chrome window while the app is running
2. **Increase timeouts** - Already set to 45 seconds, but you can increase in code if needed
3. **Check LinkedIn** - LinkedIn may be blocking automation; try:
   - Complete any security checks manually
   - Wait a few minutes between runs
   - Use a different account

### Issue 2: Checkpoint/Challenge Pages

**What it means:**
- LinkedIn detected unusual activity and requires verification
- URL contains `/checkpoint/` or `/challenge/`

**Solutions:**
1. **Complete verification manually** in the browser window
2. **Wait before running again** - Give LinkedIn time to "trust" your session
3. **Use a warmed-up account** - Log in manually a few times first
4. **Reduce automation speed** - Increase delays in `config.yaml`

### Issue 3: Search Returns 0 Profiles

**What it means:**
- No profile URLs were extracted from the search page
- Could be due to:
  - Checkpoint page blocking search
  - Page didn't load correctly
  - LinkedIn changed their HTML structure

**Solutions:**
1. **Check the browser** - Look at what page is actually displayed
2. **Complete any checkpoints** manually
3. **Try different keywords** - Some searches may have no results
4. **Increase delays** - Give pages more time to load

### Issue 4: Connection Requests Fail

**What it means:**
- Navigation to profile pages is being canceled
- Could be due to rate limiting or LinkedIn blocking

**Solutions:**
1. **Reduce daily limits** in `config.yaml` (currently 5)
2. **Increase delays** between actions (currently 3-8 seconds)
3. **Wait between runs** - Don't run multiple times in quick succession
4. **Check LinkedIn limits** - LinkedIn has weekly limits (~100 requests/week)

## Understanding the Logs

### Successful Run Example:
```
level=info msg="LinkedIn login successful"
level=info msg="running LinkedIn search" keyword="golang developer"
level=info msg="search completed" total_profiles=15
level=info msg="visiting profile to send connection request" profile="..."
level=info msg="connection request sent successfully"
level=info msg="finished processing connection requests" total_sent=5
```

### Problematic Run Example:
```
level=info msg="post‑login page URL" url="...checkpoint..."  ← Checkpoint detected
level=warning msg="failed to navigate/search, skipping keyword"  ← Search blocked
level=warning msg="context canceled"  ← Browser context lost
level=info msg="search completed" total_profiles=0  ← No profiles found
```

## Best Practices

### 1. First Time Setup
1. Run with `headless: false` to watch what happens
2. Complete any LinkedIn security checks manually
3. Let the app finish one full run
4. Check the database to see what was recorded

### 2. Regular Usage
1. **Set conservative limits** in `config.yaml`:
   ```yaml
   connect:
     daily_limit: 3  # Start small
     action_delay_min: 5s  # Longer delays
     action_delay_max: 10s
   ```

2. **Run once per day** - Don't spam LinkedIn

3. **Monitor the logs** - Watch for warnings and errors

4. **Check the database**:
   ```cmd
   sqlite3 linkedin_poc.db
   SELECT * FROM sent_requests;
   ```

### 3. When Things Go Wrong

**If you see many "context canceled" errors:**
- LinkedIn may be rate-limiting you
- Wait 24 hours before running again
- Reduce your daily limits

**If search always returns 0 profiles:**
- Check if you're logged in (look at browser window)
- Complete any checkpoint pages manually
- Try searching manually in LinkedIn first to verify it works

**If connection requests fail:**
- You may have hit LinkedIn's weekly limit
- Wait a few days before trying again
- Check LinkedIn's connection request page manually

## Configuration Tips

### Conservative Settings (Recommended)
```yaml
search:
  max_pages: 1  # Only 1 page of results
  page_delay_min: 5s
  page_delay_max: 10s

connect:
  daily_limit: 3  # Only 3 per day
  action_delay_min: 5s
  action_delay_max: 12s
```

### Aggressive Settings (Not Recommended)
```yaml
search:
  max_pages: 5  # Many pages
  page_delay_min: 1s
  page_delay_max: 2s

connect:
  daily_limit: 20  # High limit
  action_delay_min: 1s
  action_delay_max: 2s
```

## Debugging Steps

1. **Enable verbose logging** - Set `LOG_LEVEL=debug` in environment
2. **Watch the browser** - Keep `headless: false` to see what's happening
3. **Check the database** - See what was actually recorded
4. **Read the logs** - Look for patterns in errors
5. **Test manually** - Try doing the same actions manually in LinkedIn

## Getting Help

If you're stuck:
1. Check the logs for specific error messages
2. Look at what page LinkedIn is showing in the browser
3. Try reducing limits and increasing delays
4. Wait 24 hours and try again (LinkedIn may have temporarily blocked you)

## What the Code Does Now

The latest improvements:
- ✅ Detects checkpoint pages before searching
- ✅ Waits longer for pages to load (45 seconds)
- ✅ Scrolls pages to trigger lazy loading
- ✅ Better logging to show what's happening
- ✅ Graceful error handling (continues on failures)

## Expected Behavior

**Normal run:**
1. Login (may show checkpoint - complete manually)
2. Search for profiles (should find 10-50 profiles per keyword)
3. Send connection requests (respects daily limit)
4. Send follow-up messages (if connections accepted)

**If something fails:**
- App logs a warning
- Skips the failed item
- Continues with next item
- Completes successfully with partial results


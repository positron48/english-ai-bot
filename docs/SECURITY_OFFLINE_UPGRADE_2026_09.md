# Authentication and offline storage upgrade (September 2026)

- `/auth/telegram_unsafe` is retired and always returns HTTP 410. Use signed Telegram `initData` or OTP login. Signed `auth_date` is valid for 24 hours, with 60 seconds of future clock tolerance.
- Access and refresh JWTs now have distinct `token_type` claims; tokens without that claim are rejected. Existing sessions must sign in again after rollout. No database migration is required.
- Offline IndexedDB stores are namespaced by the authenticated application user ID, and course content is keyed by course. Logging out makes the old account's pending queue inaccessible; signing back into that account restores its new-format queue.
- Legacy unscoped offline stores are left intact but are not imported automatically: their owner cannot be established safely. Download offline packages again after signing in. Pending attempts from the legacy unscoped stores are not automatically synchronized.
- Removing downloaded packages preserves pending attempts in the new stores.
- Spell/type answers require `session_id` and one-based `card_index`. A retry for the last answered question returns the saved feedback while the session is in memory. Stale or unidentified answers return HTTP 409 rather than advancing the next question. Reload old clients when deploying this change.
- Transient refresh failures keep credentials. Invalid refresh tokens (HTTP 401/403) clear them. Network errors never automatically replay non-read API requests.

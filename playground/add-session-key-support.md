# Session Key Management Tab — Implementation Plan

**Objective**: Add a dedicated "Session Keys" tab to the Nitrolite Playground that allows users to register, update, revoke, and manage channel session keys through a first-class UI, plus integrate session key selection into the Channels tab for per-channel usage.

---

## Section 1: Functionality

### User-Requested Features
1. **Register session key**: Create a new session key tied to specific assets with a custom expiry duration
2. **Update session key**: Change the assets authorized or extend the expiry of an active key
3. **Revoke session key**: Immediately deactivate the current session key
4. **View session keys**: List all session keys (active and revoked) with their metadata

### Recommended Enhancements
1. **Status indicators**: Show each key's status (Active, Expired, Revoked, Expiring Soon)
2. **Version tracking**: Display the version number of each registered state for audit clarity
3. **Asset granularity**: Register different keys for different asset subsets (not just "all assets") — user selects assets via checkboxes at registration time
4. **Session key selection per channel** (in Channels tab): Let users choose which session key to use when signing for a specific channel, or fall back to wallet
5. **Auto-renewal prompts**: Warn users when a session key is within 1 hour of expiry and offer renewal
6. **Revocation history**: Maintain a local record of revoked keys with timestamps (for transparency)
7. **Multiple keys per wallet**: Support registering multiple session keys (e.g., for different apps or signing patterns)
8. **Clickable expiration display** (table column): Cycles between three formats on click — human-readable date, time-until-expiry, and raw unix timestamp. Hover shows hint for next format. Same pattern as the "Timestamp" column in HistoryTab.
9. **Copy session key address**: Allow users to copy the session key address for debugging/auditing
10. **Rate limit feedback**: If server rejects registration (e.g., max keys exceeded), show a clear error

**Rationale for enhancements**:
- **Asset granularity** is important because users may want one restricted key for a mobile app and another full-featured key for desktop. The current banner only registers "all assets," which is fine for v1 but the tab should support subsets.
- **Multiple keys per wallet** aligns with the server's design (version scoping per `(user_address, session_key, kind)` triplet), and it's common in key management UIs (e.g., SSH keys on GitHub).
- **Revocation history** is useful for debugging: "I revoked a key but did my app stop using it?" A local list (cleared on disconnect) answers that.
- **Auto-renewal prompts** reduce friction. Users appreciate a gentle nudge before lockout.
- **Per-channel selection** is a stretch goal but powerful: it lets users e.g. sign transfers with one key and deposits with another (though the v1 UI can start simpler — just a global selector).
- **Clickable expiration** avoids wasting column space on a fixed format and matches the existing HistoryTab UX pattern users already know.

---

## Section 2: UI Description

### 2.1 Session Keys Tab (New Top-Level Tab)

**Location**: Parallel to "Main" and "History" tabs in the tab bar at the top of WalletBar (only visible when wallet connected).

**Tab Name**: "Session Keys" (or "Keys" if space is tight).

**Layout**: Single card on desktop/mobile, scrollable if many keys.

#### 2.1.1 Key Management Table/List

**Container**: `SessionKeysTab.tsx` (new component)

**Columns** (table on desktop, expandable rows on mobile):
1. **Address**: Truncated session key address with copy button and icon (key icon from lucide)
2. **Assets**: Comma-separated list of asset symbols (e.g. "USDC, ETH"). Truncate/ellipsis if > 3 assets; show full list on hover in a tooltip
3. **Expiration**: Clickable value that cycles through three formats on each click:
   - `date` — human-readable UTC date/time: `2026-05-28 10:30:00 UTC`
   - `relative` — time until expiry: `23h 15m` (or `expired` if past)
   - `unix` — raw unix seconds: `1748424600`
   - Hover tooltip shows what the next click will switch to (e.g. `"Show as relative time"`, `"Show as Unix timestamp"`, `"Show as UTC date"`)
   - Default format: `relative`
   - Format state is global to the tab (one click changes all rows), same pattern as HistoryTab's `tsFormat` state
   - Styled with `.ts-toggle` CSS class (already defined in `index.css`): `cursor: pointer`, accent color on hover, tooltip via `data-tip` + CSS `::after`
   - `relative` format does **not** live-tick; it is a static snapshot rendered at mount/refresh (no `setInterval`)
4. **Status**: Pill badge (Active, Expired, Revoked, Expiring Soon, Pending). Color-coded (green=Active, gray=Expired, red=Revoked, orange=Expiring Soon, blue=Pending)
5. **Version**: Display the stored version number (e.g. "v1", "v3") for reference
6. **Actions**: Button group: [Update] [Revoke] (grayed out if already expired/revoked)

**Table styling**: Use the same card/chip styling as ChannelRow. Rows are clickable to expand detail or inline actions.

**Column filtering**: Each column header has a minimal popover for filtering (e.g. filter by status "Active"), similar to HistoryTab.

#### 2.1.2 "Register New Key" Button

**Location**: Top right of card header, next to the "Refresh" button (if a refresh button is added).

**Label**: "Register New" or "+ New Key"

**Action**: Opens an inline form or modal (`SessionKeyRegisterForm` component, new).

#### 2.1.3 Register Form (Inline or Modal)

**Title**: "Register a new session key"

**Fields**:
1. **Assets** (required, multi-select):
   - Rendered as checkboxes (not a dropdown, so users can pick subsets)
   - List all supported assets (e.g. USDC, ETH, YUSD)
   - Default: all checked
   - Validation: at least one asset must be selected
   - Display asset icons and symbols (reuse TokenSelector's asset icon logic if possible)

2. **Expiration** (required):
   - Three input modes, switchable via a small segmented control / tab: **Date**, **Duration**, **Unix**
   - **Date mode**: datetime-local input (or date + time two-field). User picks a calendar date and time. Validation: must be in the future.
   - **Duration mode**: text input with a unit selector (hours / days). Pre-filled with `24` hours as default. Displays the calculated absolute expiry below the field: `"Expires: 2026-05-28 10:30:00 UTC"`. Validation: result must be > now + 1 minute.
   - **Unix mode**: plain number input for a unix timestamp (seconds). Displays the human-readable equivalent below: `"2026-05-28 10:30:00 UTC"`. Validation: must be > now + 60.
   - All three modes resolve to the same `expires_at` (unix seconds bigint) sent to the SDK.
   - Switching between modes converts the current value where possible (e.g. switching from Duration to Date pre-fills the date with the computed value).

3. **Summary** (read-only):
   - "This key will authorize: [asset list] and expire in [duration]"
   - "On-chain operations (deposit, checkpoint, approve) will still require MetaMask"

**Buttons**:
- Cancel (dismisses form, clears state)
- Register & Sign (disabled until at least one asset is selected and expiry is valid; shows spinner during registration)

**Flow**:
1. User clicks "Register New"
2. Form opens (modal or inline below the table; modal is less disruptive for busy layouts)
3. User selects assets and expiry
4. User clicks "Register & Sign"
5. SDK is called (via `useNitrolite` client); MetaMask pops up once for wallet signature
6. Key is stored locally (already done via `useSessionKey` and `sessionKey.ts`)
7. Table refreshes and shows the new key
8. Toast: "Session key registered successfully"
9. Form closes

**Error handling**:
- If registration fails (e.g., max keys exceeded), show error toast with server message
- If user rejects MetaMask, show toast "Cancelled" (already done in `useSessionKey`)
- If network error, show "Registration failed — check your connection"

#### 2.1.4 Update Flow

**Trigger**: "Update" button on a key row

**What can be updated**:
- Assets: remove/add (keep the same private key; submit new version with updated asset list)
- Expiry: extend (set new expiry to future; version increments)

**UI**: Opens a modal similar to Register Form, but:
- Prefilled with current assets and expiry
- Title: "Update session key"
- Fields are editable
- Shows current version (e.g. "Updating from v2 to v3")
- Buttons: Cancel, Update & Sign
- Behavior is identical to registration (version check, wallet signature, etc.)

**Post-update**:
- Old key is still usable until the update is submitted; both versions exist momentarily on the server
- Updated key replaces the old one in the table (same session key address, higher version)
- Toast: "Session key updated"

#### 2.1.5 Revoke Flow

**Trigger**: "Revoke" button on a key row

**UI**: Confirmation modal
- Title: "Revoke session key"
- Message: "Are you sure? This key will no longer authorize any operations."
- Address of key being revoked (truncated, copyable)
- Buttons: Cancel, Revoke & Sign

**Backend**: Same registration flow, but `expires_at` is set to `<= now()` (e.g. current unix time), which tells the server to revoke immediately

**Post-revoke**:
- Key status changes to "Revoked"
- Buttons disabled
- Toast: "Session key revoked"
- If this was the active key in the global session, it's automatically cleared (see WalletBar integration below)

#### 2.1.6 Empty State

**When no keys exist**:
- Centered message: "No session keys yet. Register one to sign state updates without MetaMask prompts."
- Large "Register New" button

#### 2.1.7 Refresh Button

**Location**: Card header, right side (next to "Refresh" in ChannelList pattern)

**Action**: Re-fetch the list of keys from the server via `client.getLastChannelKeyStates(address, undefined, { includeInactive: true })` (already exists in SDK)

**Behavior**: Shows spinner while loading; updates the table

**Why needed**: User may revoke a key from another device; refresh brings the UI in sync.

---

### 2.2 Channels Tab Changes

**Goal**: Let users see and select which session key is active for signing operations in each channel.

**Approach** (recommended for MVP; can be simplified):
1. **Global session key selector** (simpler): One dropdown in the Channels card header that selects the active session key for all channels
   - Options: [No Key] (wallet only), [Active Key], [Other Registered Keys]
   - Display: Truncated address of selected key
   - If user switches, the SDK client is rebuilt with the new key

2. **Per-channel selector** (stretch goal, more complex): Each ChannelRow has a small session key picker
   - Allows different keys for different channels (e.g. USDC uses Key1, ETH uses Key2)
   - Requires storing a mapping in localStorage: `{ channelId -> sessionKeyAddress }`
   - Requires changes to `useChannelStates` and `useChannelOps` to pass the per-channel key choice to the signing functions
   - More flexibility but increases complexity; recommend for v2

**Recommendation for v1**: Implement the **global selector** in Channels card header. It's simpler, aligns with the current WalletBar "SK · 23h" chip design, and covers the main use case.

**UI** (global selector approach):

#### 2.2.1 Session Key Selector in ChannelList Header

**Location**: Channels card header, left side (next to the "Channels" title and count)

**Rendering**: Dropdown/select similar to the chain selector in WalletBar
- Label: "Key: " or just an icon (key from lucide)
- Selected value: Truncated address of active key, or "Wallet" if none
- Options list:
  - "Wallet (no key)" — always first, forces re-build of client with wallet signer only
  - Separator
  - [List of all non-expired keys, grouped by status]
  - Separator
  - [List of expired/revoked keys, grayed out]

**Example**:
```
Key: 0x1a2B…c3dE
  v
[ Wallet (no key)
  ──────────────
  ✓ 0x1a2B…c3dE (active, expires 23h)
    0x9fEd…A1b2 (expires 2h) ← orange badge "Expiring Soon"
  ──────────────
  0xAbCd…1234 (expired)
]
```

**On selection change**:
1. Call `sk.selectKey(selectedAddress)` or similar in useSessionKey (new action)
2. Rebuild the SDK Client in `useNitrolite` with the new key (if a key is selected) or wallet signer (if "Wallet" is selected)
3. Clear old key from session state if a different one is selected
4. Toast: "Switched to [address]" or "Using wallet only"

#### 2.2.2 Visual Indicator

**In ChannelRow**: Small badge or icon next to the channel name indicating which key is active
- Example: Green key icon if an SK is active, or nothing if wallet only
- Tooltip on hover: "Signed with: [key address] / Wallet"

**Rationale**: Users should know at a glance whether they're using a key or wallet for a channel.

---

### 2.3 WalletBar / Global Session Key Indicator

**Current state**: Already has a chip showing active SK with expiry countdown and clear button.

**Changes**:
1. Update the chip to link to the Session Keys tab (e.g., click the chip to switch to "Session Keys" tab, or add a "Manage" link next to the chip)
2. If a key expires while the tab is active, refresh the chip in real-time (already done via `setInterval` in `useSessionKey`)
3. If a key is revoked from the server (detected on next refresh), clear it from state and show banner again

**No major changes needed**; the current design is good.

---

### 2.4 Action Panel Changes

**Current state**: Deposit, withdraw, transfer, faucet tabs; all signing is invisible (handled by `useChannelOps`).

**Changes**: None required for v1. The Client already uses the active session key (or wallet) transparently.

**Future enhancement** (v2):
- Show a badge on the Action Panel indicating which key is signing (e.g. "Signing with SK: 0x1a2B…c3dE" or "Signing with wallet")
- Allows users to switch keys mid-operation without re-selecting the asset

---

## Section 3: Implementation Plan

### 3.1 Files to Create

#### 1. `src/components/SessionKeysTab.tsx`
**Purpose**: Main tab UI, displays list of session keys and register form.

**Responsibilities**:
- Fetch list of keys from server on mount (via `client.getLastChannelKeyStates`)
- Render table/list with columns (address, assets, expires, status, version, actions)
- Render "Register New" button and form modal
- Handle register/update/revoke flows
- Manage local revocation history (Set of revoked addresses) for quick visual feedback
- Refresh action

**State**:
- `keys: ChannelSessionKeyStateV1[]` — list from server
- `showRegisterModal: boolean` — toggle form modal
- `selectedKeyForUpdate: ChannelSessionKeyStateV1 | null` — for update flow
- `isLoading: boolean` — while fetching keys
- `isSubmitting: boolean` — while registering/updating/revoking
- `revokedLocally: Set<Address>` — addresses revoked in this session (cleared on disconnect)

**Props**:
- `client: Client | null`
- `address: Address | null`
- `sessionKey: StoredSessionKey | null` (from App.tsx)
- `onSessionKeyChanged: (key: StoredSessionKey | null) => void` (triggers client rebuild in App)
- `supportedAssets: Asset[]`

#### 2. `src/components/SessionKeyRegisterForm.tsx`
**Purpose**: Reusable form for registering and updating session keys.

**Responsibilities**:
- Render asset checkboxes, expiry picker, summary
- Validate inputs
- Call registration/update flow via `useSessionKey` or new hook
- Handle loading/error states

**Props**:
- `mode: 'register' | 'update'`
- `initialAssets?: string[]` (for update mode)
- `initialExpirySeconds?: number` (for update mode)
- `supportedAssets: Asset[]`
- `isSubmitting: boolean`
- `onSubmit: (assets: string[], expirySeconds: number) => Promise<void>`
- `onCancel: () => void`

#### 3. `src/hooks/useSessionKeyManagement.ts`
**Purpose**: High-level session key operations beyond basic registration/clearing.

**Responsibilities**:
- Fetch list of keys from server
- Handle registration flow (calls existing `useSessionKey.register`)
- Handle update flow (version increment, re-sign, submit)
- Handle revoke flow (version increment, expires_at = now, submit)
- Track loading/error states per operation

**Interface**:
```typescript
interface UseSessionKeyManagementResult {
  keys: ChannelSessionKeyStateV1[];
  isLoading: boolean;
  isSubmitting: boolean;
  fetchKeys: () => Promise<void>;
  register: (assets: string[], expirySeconds: number) => Promise<void>;
  update: (key: ChannelSessionKeyStateV1, assets: string[], expirySeconds: number) => Promise<void>;
  revoke: (key: ChannelSessionKeyStateV1) => Promise<void>;
}

function useSessionKeyManagement(
  client: Client | null,
  address: Address | null,
): UseSessionKeyManagementResult
```

---

### 3.2 Files to Modify

#### 1. `src/App.tsx`
**Changes**:
- Add "Session Keys" to `AppTab` type: `type AppTab = 'main' | 'history' | 'keys'`
- Update WalletBar to show the new tab button
- Render `<SessionKeysTab>` when `activeTab === 'keys'`
- Pass `onSessionKeyChanged` callback to SessionKeysTab to allow it to select a different key globally

**Code snippet** (pseudo):
```typescript
type AppTab = 'main' | 'history' | 'keys';

<WalletBar
  // ... existing props
  activeTab={activeTab}
  onTabChange={setActiveTab}
/>

// In main body:
{activeTab === 'history' ? (
  <HistoryTab ... />
) : activeTab === 'keys' ? (
  <SessionKeysTab
    client={nitro.client}
    address={wallet.address}
    sessionKey={sk.sessionKey}
    supportedAssets={nitro.supportedAssets}
    onSessionKeyChanged={(key) => {
      if (key) sk.sessionKey is already set, but allow switching
      // Actually, switching keys = selecting different key to use for signing
      // This requires a new action in useSessionKey: selectKey(address)
    }}
  />
) : (
  <div className="grid ...">
    <ActionPanel ... />
    <ChannelList ... />
  </div>
)}
```

#### 2. `src/components/WalletBar.tsx`
**Changes**:
- Add "Session Keys" tab button to the center tab selector (only when wallet connected)
- Make the session key chip clickable to navigate to Session Keys tab (optional enhancement)

**Code snippet** (pseudo):
```typescript
<button
  className={`tab${activeTab === 'keys' ? ' active' : ''}`}
  onClick={() => onTabChange('keys')}
>
  Session Keys
</button>
```

#### 3. `src/hooks/useSessionKey.ts`
**Changes**:
- Add new action: `selectKey(sessionKeyAddress: Address) -> void`
  - Finds the stored key by address in localStorage
  - Sets it as the active session key (triggers re-render)
  - Used when user selects a different key from the Channels dropdown

- Add new action: `getAllStoredKeys() -> StoredSessionKey[]`
  - Scans localStorage for all keys matching the node URL and wallet address (e.g., multiple versions of the same key)
  - Returns all non-expired keys (or optionally includes expired for the Session Keys tab)

**Updated interface**:
```typescript
export interface UseSessionKeyResult {
  sessionKey: StoredSessionKey | null;
  allKeys: StoredSessionKey[];  // NEW: all registered keys for this wallet
  isRegistering: boolean;
  register: (client: Client, assetSymbols: string[]) => Promise<void>;
  selectKey: (sessionKeyAddress: Address) => void;  // NEW: switch to a different key
  clear: () => void;
}
```

#### 4. `src/hooks/useNitrolite.ts`
**Changes**:
- Update Client rebuild logic to watch for changes in the selected session key (already tracks `sessionKey` prop, but may need to support selecting a different key without clearing the old one)
- When `sessionKey` prop changes (or new `selectKey` is called), rebuild the client with the new key

**No major changes needed**; the existing dependency on `sessionKey` in the useEffect should handle this.

#### 5. `src/components/ChannelList.tsx`
**Changes**:
- Add session key selector dropdown in the card header (global selector for v1)
- Pass `onSessionKeySelected` callback to SessionKeysTab or accept it as a prop and expose a dropdown

**Code snippet** (pseudo):
```typescript
<div className="card-header">
  <div className="flex items-center gap-2">
    <span className="card-title">Channels</span>
    <span className="text-text-muted text-sm mono">({channels.length})</span>
    
    {/* NEW: Session Key Selector */}
    <select
      value={activeSessionKeyAddress ?? 'wallet'}
      onChange={(e) => {
        if (e.target.value === 'wallet') sk.clear();
        else sk.selectKey(e.target.value as Address);
      }}
      className="chip mono text-xs cursor-pointer"
    >
      <option value="wallet">Wallet (no key)</option>
      {/* List keys */}
    </select>
  </div>
  {/* Refresh button */}
</div>
```

#### 6. `src/sessionKey.ts`
**Changes**:
- No major changes. Existing `registerSessionKey` can be reused for registration and updates (the caller just increments the version before calling)
- Add helper function: `getStoredSessionKeys(nodeUrl: string, walletAddress: Address) -> StoredSessionKey[]`
  - Returns all non-expired keys from localStorage for this wallet (or add `includeExpired` param)

#### 7. `src/utils.ts`
**Changes**:
- Add helper: `formatSessionKeyAddress(address: Address, length: number = 4) -> string`
  - Returns truncated address like "0x1a2B…c3dE"
- Add helper: `formatExpiryCountdown(expiresAt: number) -> string`
  - Returns "23h", "2h 15m", "45m", "15s", "expired"
  - Used in SessionKeysTab and WalletBar chip

#### 8. `REFERENCE.md`
**Changes**:
- Add new section: "### SessionKeysTab" with description of purpose, state, and user flows
- Update "### WalletBar" to mention the Session Keys tab button and global key selector in ChannelList
- Update "## Key User Flows" to add new flow: "Set up multiple session keys" or "Switch between session keys"

**New entry** (pseudo):
```markdown
### SessionKeysTab
**For**: Registering, managing, and selecting session keys for signing state operations.
- List all registered keys (active, expired, revoked) with assets, expiry, and version
- "Register New" button opens a form to pick assets and expiry duration
- "Update" button allows changing assets or extending expiry (version increment)
- "Revoke" button deactivates a key immediately
- Global session key selector in ChannelList header to choose which key signs for all channels

### ChannelList Changes
- Added session key selector dropdown (global, for v1)
- Allows switching between active session keys or wallet-only mode
- Rebuilds SDK Client when selection changes
```

---

### 3.3 New Dependencies / SDK Calls

**No new npm packages required.** The implementation uses:
- Existing `@yellow-org/sdk` methods: `getLastChannelKeyStates`, `submitChannelSessionKeyState`
- Existing React hooks and Tailwind CSS
- `lucide-react` icons (already in use)

---

### 3.4 State Management Summary

| State | Owner | Scope | Notes |
|-------|-------|-------|-------|
| Active session key | `useSessionKey` | Wallet + node | Persisted in localStorage; used to build Client |
| All stored keys | `useSessionKey` (new) | Wallet + node | Non-expired keys only, for selector dropdown |
| Keys from server | `useSessionKeyManagement` (new) | Session only | Fetched via `getLastChannelKeyStates`; includes inactive |
| Session key selector choice | `App.tsx` (or `useSessionKey`) | Session only | Which key is actively signing; impacts Client rebuild |
| Register form visibility | `SessionKeysTab` | Component | Local state for modal |
| Revoked locally | `SessionKeysTab` | Session only | Set of addresses revoked in this session for instant UI feedback |

---

### 3.5 Order of Implementation

**Recommended sequence**:

1. **Phase 1 — Core UI**:
   - Create `SessionKeysTab.tsx` (minimal: list only, no forms)
   - Create `useSessionKeyManagement.ts` hook for fetching keys
   - Update `App.tsx` to add "Session Keys" tab
   - Update `WalletBar.tsx` to show the tab button
   - Update `REFERENCE.md` with new components

2. **Phase 2 — Register/Update/Revoke Forms**:
   - Create `SessionKeyRegisterForm.tsx`
   - Add form modal to `SessionKeysTab`
   - Implement register, update, revoke flows (call SDK methods, handle errors, toasts)
   - Test with actual Nitronode

3. **Phase 3 — Key Selection**:
   - Update `useSessionKey.ts` to add `selectKey` and `getAllStoredKeys` actions
   - Add session key selector dropdown to `ChannelList` header
   - Wire up Client rebuild when key selection changes
   - Test switching keys mid-session

4. **Phase 4 — Polish**:
   - Add status badges (Active, Expired, Revoked, Expiring Soon)
   - Add column filtering in SessionKeysTab (optional, MVP may skip)
   - Add per-channel session key indicator (optional, stretch goal)
   - Add auto-renewal prompts when key is < 1 hour from expiry
   - Add copy buttons and tooltips

---

## Section 4: Technical Notes

### 4.1 Version Increment Logic

When registering, updating, or revoking a key:
1. Call `client.getLastChannelKeyStates(address, undefined, { includeInactive: true })`
2. Find the highest `version` among results for this `(session_key, kind)` pair
3. Increment to `nextVersion = highest + 1` (or `1` if no results)
4. Submit with `version: nextVersion.toString()`
5. If the server rejects with "expected version N, got M", it means a race condition occurred (another tab/device registered a version). Retry step 1-4.

The `registerSessionKey` function in `sessionKey.ts` already does this; reuse it for update/revoke too.

### 4.2 Expiry and Renewal

- **Expiry time**: Unix seconds (bigint)
- **Renewal buffer**: 5 minutes (already in `sessionKey.ts`)
- **Client-side check**: `isExpired(sk)` returns true if `expiresAt - 5min <= now`
- **Server-side check**: `expiresAt > now` required at transaction time
- **Auto-renewal**: When key is < 1 hour from expiry, show a banner or button in SessionKeysTab; clicking "Renew" opens the Update form with the same assets pre-filled and Duration mode defaulting to `24 hours`

### 4.6 Expiration Column Formatting

Three-way cycling format, global to the SessionKeysTab component (one `expFmt` state, same pattern as `tsFormat` in HistoryTab):

```typescript
type ExpFormat = 'relative' | 'date' | 'unix';

function formatExpiry(expiresAt: number, fmt: ExpFormat): string {
  const now = Math.floor(Date.now() / 1000);
  if (fmt === 'unix') return expiresAt.toString();
  if (fmt === 'date') return new Date(expiresAt * 1000).toISOString().replace('T', ' ').slice(0, 19) + ' UTC';
  // relative
  const diff = expiresAt - now;
  if (diff <= 0) return 'expired';
  const h = Math.floor(diff / 3600);
  const m = Math.floor((diff % 3600) / 60);
  return h > 0 ? `${h}h ${m}m` : `${m}m`;
}

const NEXT_FMT: Record<ExpFormat, ExpFormat> = {
  relative: 'date',
  date: 'unix',
  unix: 'relative',
};

const FMT_HINT: Record<ExpFormat, string> = {
  relative: 'Show as UTC date',
  date: 'Show as Unix timestamp',
  unix: 'Show as relative time',
};
```

The clickable `<span>` uses the existing `.ts-toggle` CSS class from `index.css` (cursor pointer, accent color on hover, tooltip via `data-tip` attribute + `::after` pseudo-element). No new CSS needed.

### 4.3 localStorage Key Scheme

Current: `nitrolite_playground_sk_{nodeUrl}::{walletAddress}` (single key per wallet per node)

For multiple keys: Consider extending to `nitrolite_playground_sk_{nodeUrl}::{walletAddress}::{sessionKeyAddress}` to store each version/address separately. Alternatively, store a JSON array of keys under a single localStorage key.

**Recommendation**: Use an array approach: `nitrolite_playground_sk_{nodeUrl}::{walletAddress}` = `[{ sessionKeyAddress, privateKey, ... }, ...]`. This way:
- Single localStorage entry per wallet per node
- Easy to iterate and filter
- Existing `loadSessionKey` becomes `loadActiveSessionKey` (returns the one in use); new `loadAllSessionKeys` returns the array

### 4.4 Error Handling

**Common errors from server**:
- `invalid_session_key_state: expected version N, got M` → Retry with correct version (race condition)
- `session key does not have permission to sign for this data` → Asset not in registered list; update key
- `max_session_keys_exceeded` → User hit server's per-wallet limit; show error and suggest revoking an old key
- `Token signature invalid` → Shouldn't happen if SDK is correct; check key freshness

**Client-side validation**:
- Expiry must be in the future (at least 1 minute)
- At least one asset must be selected
- Form disables submit until valid

### 4.5 Testing Checklist

- [ ] Register a session key with all assets, 24h expiry
- [ ] Register a second key with subset of assets (e.g., USDC only), 7 days expiry
- [ ] Update the first key to remove an asset
- [ ] Extend the second key's expiry
- [ ] Revoke the first key (status changes to "Revoked", buttons disabled)
- [ ] Switch to the second key via ChannelList dropdown; perform a deposit (should use second key for signing)
- [ ] Switch back to wallet-only; perform a withdraw (should use MetaMask for signing)
- [ ] Let a key expire in real-time; verify status badge updates every second
- [ ] Disconnect and reconnect wallet; verify correct key is loaded from localStorage
- [ ] Manually modify localStorage to corrupt a key; verify app handles it gracefully (clears key, shows error or banner)
- [ ] Revoke a key from another device/session; refresh SessionKeysTab; verify status updates

---

## Section 5: Future Enhancements

1. **Per-channel session key assignment**: Allow selecting different keys for different channels (requires mapping in localStorage, changes to signing logic)
2. **Cloud backup**: Encrypt and backup session keys to a secure server (requires new backend, key derivation, security review)
3. **Session key templates**: Pre-configured keys (e.g., "Mobile Read-Only: USDC only, 7d expiry") for common patterns
4. **Key rotation automation**: Schedule a key to be rotated/renewed every N days without user intervention
5. **Audit log**: Per-key log of operations signed (requires server support)
6. **Hardware wallet integration**: Support signing session key ownership with hardware wallet (Ledger, Trezor)
7. **Key expiry notifications**: Email/push notif when key expires or is revoked from another device

---

## Summary

This plan delivers a production-ready "Session Keys" tab that gives users fine-grained control over session key registration, updates, revocation, and selection. The UI is consistent with the existing Nitrolite design, the implementation reuses existing SDK capabilities, and the phased approach allows for MVP launch followed by polish and stretch goals.

**MVP Scope** (Phase 1 + 2):
- Session Keys tab with list and register form
- Update and revoke flows
- Integration with existing session key storage

**Nice-to-Have** (Phase 3 + 4):
- Per-channel key selection
- Status badges and auto-renewal prompts
- Multiple keys per wallet support

---

## Section 6: Mockup Specifications

Mockups live in `mockups/session-keys/`. Each file is a self-contained HTML + CSS document using the exact design tokens, fonts, and class patterns from the live app (`src/index.css`). Use these as the pixel-level reference for implementation.

---

### 6.1 `mockup-1-key-list.html` — Session Keys tab, key list view

**What it shows**: The steady-state Session Keys tab with four keys in all possible statuses.

#### WalletBar changes
- "Session Keys" added as the **third tab** in the center tab group: `[Main] [History] [Session Keys]`
- Tab uses the same `.tab` / `.tab.active` pill classes as the existing tabs
- The existing SK chip in the top-right WalletBar (`SK · 23h ×`) remains unchanged; it still shows the active key and expiry

#### Expiring Soon banner
- Rendered above the card when any registered key is within the warning threshold of expiry
- Layout: orange icon circle (32px, `rgba(249,115,22,0.18)` background with orange key SVG) + text block + "Renew Key" ghost button
- Container: `border-radius: 10px`, `background: var(--warning-dim)` (`rgba(249,115,22,0.12)`), `border: 1px solid rgba(249,115,22,0.25)`, `margin-bottom: 16px`
- Text: title in `#f97316`, subtitle in `var(--text-muted)` naming the key address (truncated mono) and time remaining
- "Renew Key" button: ghost style with `border-color: rgba(249,115,22,0.4)` and `color: #f97316`

#### Card header
- Left: key SVG icon (muted color) + "Session Keys" `.card-title` + count badge (same rounded pill style as ChannelList — `bg-elevated`, `border`, muted text)
- Right: "Refresh" ghost button (icon + label) + "Register New" primary button (+ icon + label)
- Buttons use `.btn .btn-ghost .btn-sm` and `.btn .btn-primary .btn-sm` respectively

#### Table columns

| Column | Header style | Cell content |
|--------|-------------|--------------|
| **Address** | Plain `.th-inner` | Key icon SVG (colored by status) + `JetBrains Mono` 12px truncated address + optional "IN USE" chip (accent-dim background, `font-size:10px`, bold) for the currently active key + copy button (`.copy-btn-addr`, appears on row hover) |
| **Assets** | Plain `.th-inner` | Comma-separated asset symbols in `JetBrains Mono` 12px muted |
| **Expiration** | `.th-inner.filterable` with chevron | `.ts-toggle` span cycling `relative → date → unix` on click; `data-tip` shows next format name. Expired = "expired" in muted; expiring soon = value in `#f97316` |
| **Status** | `.th-inner.filterable` with chevron | Pill badge (see below) |
| **Version** | Plain `.th-inner` | `JetBrains Mono` 12px muted, e.g. "v1", "v3" |
| **Actions** | Right-aligned plain | Button group flush-right (see below) |

#### Status badges
All use `.badge` base class (inline-flex, `border-radius: 999px`, `padding: 3px 9px`, `font-size: 11px`, `font-weight: 500`), plus a colored dot prefix:

| Status | Class | Background | Text color |
|--------|-------|------------|------------|
| Active | `.badge-active` | `rgba(34,197,94,0.12)` | `#22c55e` |
| Expiring Soon | `.badge-expiring` | `rgba(249,115,22,0.14)` | `#f97316` |
| Expired | `.badge-expired` | `rgba(102,102,102,0.14)` | `var(--text-muted)` |
| Revoked | `.badge-revoked` | `rgba(239,68,68,0.12)` | `var(--error)` |
| Pending | `.badge-pending` | `rgba(59,130,246,0.12)` | `#3b82f6` |

#### Action buttons per row

| Row status | Left button | Right button |
|-----------|-------------|--------------|
| Active | `.btn .btn-ghost .btn-sm` "Update" | `.btn .btn-danger .btn-sm` "Revoke" |
| Expiring Soon | Styled ghost `border-color:rgba(249,115,22,0.4); color:#f97316` "Renew" | `.btn .btn-danger .btn-sm` "Revoke" |
| Expired | `.btn .btn-ghost .btn-sm` disabled | `.btn .btn-danger .btn-sm` disabled |
| Revoked | `.btn .btn-ghost .btn-sm` disabled | `.btn .btn-danger .btn-sm` disabled |

#### Row opacity
- Active / Expiring Soon: 100%
- Expired: `opacity: 0.7`
- Revoked: `opacity: 0.55`

#### Row hover
- `tbody tr:hover { background: rgba(255,255,255,0.02); }`

---

### 6.2 `mockup-2-register-modal.html` — Register New Key modal

**What it shows**: The register modal open over a blurred Session Keys tab, in "Duration" expiry mode with USDC + ETH selected.

#### Background
- Session Keys tab is rendered at `filter: blur(1.5px); opacity: 0.35; pointer-events: none` to give context without distraction

#### Overlay
- `.modal-overlay`: `position: fixed; inset: 0; background: rgba(0,0,0,0.72); backdrop-filter: blur(4px); z-index: 50`

#### Modal card
- `.modal-card` + custom: `width: 440px; max-width: calc(100vw - 32px); padding: 28px; display: flex; flex-direction: column; gap: 20px`
- Sections are separated by full-bleed horizontal hairlines: `height: 1px; background: var(--border); margin: 0 -28px`

#### Header section (centered)
- 48px circle: `background: var(--accent-dim)` + key SVG in `var(--accent)`
- Title: 17px, `font-weight: 600`, centered
- Subtitle: 13px, `var(--text-muted)`, centered, 1.5 line-height

#### Assets section
- Section label: 12px, `font-weight: 500`, `var(--text-muted)`, uppercase, `letter-spacing: 0.05em`
- "Select all · Clear" right-aligned in 11px muted, cursor pointer
- Each asset row: flex row, `padding: 10px 14px`, `border-radius: 8px`, `background: var(--bg-elevated)`, `border: 1px solid var(--border)`
  - **Unchecked**: default style above
  - **Checked**: `border-color: var(--accent); background: var(--accent-dim)`
- Checkbox box: 16px × 16px, `border-radius: 4px`, `border: 2px solid var(--border)`; when checked: `background: var(--accent); border-color: var(--accent)` + white checkmark SVG (stroke-width 3.5)
- Asset icon: 24px circle with brand color background; asset symbol + full name beside it (13px medium + 11px muted)

#### Expiry section
- Segmented control: outer `background: var(--bg-elevated); border: 1px solid var(--border); border-radius: 8px; padding: 3px; gap: 2px`
- Each segment button: `border-radius: 6px; font-size: 12px; font-weight: 500; color: var(--text-muted)`; active = `background: var(--bg-surface); border: 1px solid var(--border); color: var(--text-primary)`
- **Duration mode** (default): number `.input-wrap` (flex, `bg-elevated`, `border`, `border-radius: 8px`) containing plain number input (`JetBrains Mono` 14px) + `<select>` (hours / days) with `border-left: 1px solid var(--border); background: var(--bg-base); padding: 0 12px`
- Computed timestamp line below: 12px muted "Expires: " + mono inline span for the timestamp value
- **Date mode**: replaces the duration row with a `datetime-local` input in the same `.input-wrap` style
- **Unix mode**: replaces with a plain number input + mono human-readable equivalent below

#### Summary box
- `background: var(--bg-elevated); border: 1px solid var(--border); border-radius: 8px; padding: 12px 14px; font-size: 12px; color: var(--text-muted); line-height: 1.6`
- Key asset names and expiry duration rendered in `<strong>` (`color: var(--text-primary); font-weight: 500`)

#### Footer
- Two equal-flex buttons side by side: Cancel (`.btn .btn-ghost`) + "Register & Sign" (`.btn .btn-primary`)
- During submission: "Register & Sign" shows `.spinner` + "Signing…"

---

### 6.3 `mockup-3-channels-integration.html` — Channels tab, per-channel SK selector

**What it shows**: The Main tab in its normal two-column layout, with a session key selector surfaced inside each opened channel's expanded panel. The mockup renders a dropdown in open state for reference.

> **Design correction**: The mockup renders the SK selector in the Channels card header for illustration of the dropdown anatomy. The actual implementation must place the selector **inside each opened ChannelRow's expanded panel** (the section revealed by clicking the row), not in the card header. Closed channels must not show a selector. This matches the constraint that session key selection applies only to channels where state-update signing is relevant.

#### Main tab layout (unchanged)
- Two-column grid: `grid-template-columns: 420px 1fr; gap: 20px; align-items: start`
- Left: `ActionPanel` card (Deposit / Withdraw / Transfer / Faucet sub-tabs)
- Right: `ChannelList` card

#### ChannelList card header (unchanged from current)
- Left: "Channels" title + `(N)` count in `.text-text-muted .text-sm .mono`
- Right: Refresh icon-only button (`.btn .btn-ghost .btn-sm`)
- No global SK selector in the header

#### SK selector inside ChannelRow expanded panel (new)
- Rendered in the expanded section (`border-t border-border px-4 py-3`) **only when `channel.status !== ChannelStatus.Closed`**
- Positioned alongside the existing "Close channel" button (same row or directly above it)
- Trigger button: same `.chip`-style as other inline controls — shows active key icon (colored dot) + truncated mono address (or "Wallet" if none), + chevron
- Trigger label variants:
  - No key selected: gray wallet icon + "Wallet"
  - Active key: green dot + `0x1a2B…c3dE`
  - Expiring key: orange dot + `0x9fEd…A1b2`

#### SK dropdown anatomy (open state)
- Container: `background: var(--bg-elevated); border: 1px solid var(--border); border-radius: 10px; box-shadow: 0 8px 24px rgba(0,0,0,0.55); min-width: 260px; padding: 6px 0`
- Items use `padding: 9px 14px; font-size: 13px`, hover = `rgba(255,255,255,0.04)`, selected = `var(--accent-dim)`
- **Wallet option**: wallet icon + "Wallet (no key)" label + right-aligned "MetaMask" muted tag
- **Separator**: `height: 1px; background: var(--border); margin: 4px 0`
- **Active key item**: green 6px dot + mono address + right-aligned time-remaining + checkmark SVG for selected item
- **Expiring key item**: orange 6px dot + mono address + right-aligned time in `#f97316` + `⚠` glyph
- **Expired key item**: `opacity: 0.45; cursor: not-allowed` — not clickable
- **Bottom separator + "Manage session keys →"** link item: accent color, navigates to Session Keys tab

#### Per-channel SK badge in collapsed row header (read-only indicator)
- Small inline badge next to the channel name, always visible (collapsed and expanded)
- **When SK is active**: `background: rgba(34,197,94,0.1); border: 1px solid rgba(34,197,94,0.25); color: var(--success)` + 9px key SVG + truncated address, e.g. "SK · 0x1a2B…c3dE"
- **When wallet only**: `background: var(--bg-elevated); border: 1px solid var(--border); color: var(--text-muted)` + wallet SVG + "Wallet"
- Badge uses `border-radius: 5px; padding: 2px 7px; font-size: 10px; font-weight: 500`


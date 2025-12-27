# Euro Support Migration Plan

## Goal

Add Euro (EUR) support alongside Bulgarian Lev (BGN) while maintaining 100% backward compatibility. Frontend should work with zero changes after backend updates.

## Current State

- All expenses are in BGN (Leva) - no currency field existed
- Added `is_euro` boolean to expenses table (default: false)
- Exchange rate: 1 EUR = 1.95583 BGN (fixed rate)

## Design Principles

1. **Backward compatible** - no breaking changes to existing API responses
2. **Default to Leva** - if no currency info provided, assume BGN
3. **Per-expense currency** - each expense knows its own currency
4. **Global preference** - users can set a default currency for new expenses
5. **Optional conversion** - API can return amounts in both currencies

---

## Implementation Plan

### Phase 1: Per-Expense Currency (Already Done)

- [x] Add `is_euro` boolean to Expense struct
- [x] Add `is_euro` to SELECT queries
- [x] Add `is_euro` to INSERT (defaults to false if not provided)
- [x] Migration file: `001_add_is_euro_to_expenses.sql`

### Phase 2: User Currency Preference

Add a default currency preference per user so new expenses auto-use their preferred currency.

**Migration:** `002_add_default_currency_to_users.sql`
```sql
ALTER TABLE users ADD COLUMN default_currency VARCHAR(3) NOT NULL DEFAULT 'BGN';
```

**API Changes:**
- `GET /api/v1/me` or `GET /api/v1/user/preferences` - returns user's default currency
- `PUT /api/v1/user/preferences` - update default currency

**Backward Compatibility:**
- Existing users get `default_currency = 'BGN'` by default
- Frontend doesn't need to call these endpoints - everything works as before

### Phase 3: Enhanced Expense Response (Optional)

Add computed fields to expense responses for convenience. These are **additional** fields, not replacements.

**Option A: Add converted amount to response**
```json
{
  "id": 1,
  "amount": 100.00,
  "is_euro": false,
  "amount_bgn": 100.00,
  "amount_eur": 51.13
}
```

**Option B: Add currency code string**
```json
{
  "id": 1,
  "amount": 100.00,
  "is_euro": false,
  "currency": "BGN"
}
```

**Backward Compatibility:**
- Original `amount` and `is_euro` fields remain unchanged
- New fields are additive - frontend ignores what it doesn't use

### Phase 4: Query Filtering by Currency (Optional)

Allow filtering expenses by currency.

**New query parameter:**
```
GET /api/v1/expenses?currency=EUR
GET /api/v1/expenses?currency=BGN
GET /api/v1/expenses?currency=all  (default, current behavior)
```

**Backward Compatibility:**
- No parameter = return all expenses (current behavior)

### Phase 5: Aggregates with Currency Awareness (Optional)

When using `aggregates_only=true` or `group_by`, handle mixed currencies.

**Option A: Separate totals**
```json
{
  "total_bgn": 1500.00,
  "total_eur": 200.00,
  "total_bgn_equivalent": 1891.17
}
```

**Option B: Convert everything to one currency**
```
GET /api/v1/expenses?aggregates_only=true&convert_to=BGN
```

---

## Recommended Minimum Implementation

For seamless frontend compatibility with future Euro support:

### Must Have (Phase 1-2)
1. ✅ `is_euro` field on expenses (done)
2. `default_currency` on users table
3. Create expense uses user's default currency if `is_euro` not specified
4. Return `is_euro` in expense responses (already done)

### Nice to Have (Phase 3+)
5. Add `currency` string field to response ("EUR" or "BGN")
6. Add `amount_bgn` and `amount_eur` computed fields
7. Currency filter on expense queries

---

## API Behavior Summary

| Scenario | Behavior |
|----------|----------|
| Create expense, no `is_euro` field | Use user's `default_currency` (falls back to BGN) |
| Create expense, `is_euro: true` | Store as Euro |
| Create expense, `is_euro: false` | Store as Leva |
| Get expenses | Returns `is_euro` field (frontend decides display) |
| Existing expenses | All have `is_euro: false` (Leva) |
| Old frontend (no currency awareness) | Works perfectly - ignores `is_euro` field |

---

## Files to Modify

| File | Changes |
|------|---------|
| `migrations/002_add_default_currency_to_users.sql` | Add column to users |
| `database.go` | Add `DefaultCurrency` to User struct |
| `handlers.go` | Update create expense to use user's default currency |
| `handlers.go` | Add user preferences endpoint (optional) |

---

## Conversion Helper

Add to `database.go` or new `currency.go`:

```go
const EUR_TO_BGN = 1.95583

func convertToBGN(amount float64, isEuro bool) float64 {
    if isEuro {
        return amount * EUR_TO_BGN
    }
    return amount
}

func convertToEUR(amount float64, isEuro bool) float64 {
    if !isEuro {
        return amount / EUR_TO_BGN
    }
    return amount
}
```

---

## Testing Checklist

- [ ] Create expense without `is_euro` → defaults to user preference or BGN
- [ ] Create expense with `is_euro: true` → stored as Euro
- [ ] Create expense with `is_euro: false` → stored as Leva
- [ ] Get expenses returns `is_euro` for each
- [ ] Old frontend works without changes
- [ ] Aggregates still work (even with mixed currencies)
- [ ] Existing data unchanged after migration

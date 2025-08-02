## Expenses

All expenses: GET /api/v1/expenses
User 1 expenses: GET /api/v1/expenses?user_id=1
User 2 expenses: GET /api/v1/expenses?user_id=2
NULL user expenses: GET /api/v1/expenses?user_id=0
Total for all: GET /api/v1/expenses?aggregates_only=true
Total for user 1: GET /api/v1/expenses?user_id=1&aggregates_only=true
Total for NULL users: GET /api/v1/expenses?user_id=0&aggregates_only=true
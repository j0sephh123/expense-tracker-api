## Expenses

All expenses: GET /api/v1/expenses
User 1 expenses: GET /api/v1/expenses?user_id=1
User 2 expenses: GET /api/v1/expenses?user_id=2
NULL user expenses: GET /api/v1/expenses?user_id=0
Total for all: GET /api/v1/expenses?aggregates_only=true
Total for user 1: GET /api/v1/expenses?user_id=1&aggregates_only=true
Total for NULL users: GET /api/v1/expenses?user_id=0&aggregates_only=true

GET /api/v1/expenses?category_id=1 - expenses from category 1
GET /api/v1/expenses?category_id=1,2 - expenses from categories 1 or 2
GET /api/v1/expenses?user_id=1&category_id=2 - expenses for user 1 from category 2
GET /api/v1/expenses?category_id=1,2&aggregates_only=true - total amount for categories 1 and 2

GET /api/v1/expenses?subcategory_id=4 - expenses from subcategory 4
GET /api/v1/expenses?subcategory_id=4,5 - expenses from subcategories 4 or 5
GET /api/v1/expenses?user_id=1&subcategory_id=4 - expenses for user 1 from subcategory 4
GET /api/v1/expenses?subcategory_id=4,5&aggregates_only=true - total amount for subcategories 4 and 5

GET /api/v1/expenses?order_by=amount&order_dir=desc - expenses ordered by amount (highest first)
GET /api/v1/expenses?order_by=amount&order_dir=asc - expenses ordered by amount (lowest first)
GET /api/v1/expenses?order_by=date&order_dir=desc - expenses ordered by date (newest first)
GET /api/v1/expenses?order_by=date&order_dir=asc - expenses ordered by date (oldest first)
GET /api/v1/expenses?user_id=1&order_by=amount&order_dir=desc - user 1 expenses by amount (highest first)

GET /api/v1/expenses?date_from=2025-07-01 - expenses from July 1, 2025 onwards
GET /api/v1/expenses?date_to=2025-07-31 - expenses up to July 31, 2025
GET /api/v1/expenses?date_from=2025-07-01&date_to=2025-07-31 - expenses in July 2025
GET /api/v1/expenses?user_id=1&date_from=2025-07-01 - user 1 expenses from July 1, 2025
GET /api/v1/expenses?date_from=2025-07-01&date_to=2025-07-31&aggregates_only=true - total amount for July 2025

GET /api/v1/expenses?group_by=category - expenses grouped by category
GET /api/v1/expenses?group_by=subcategory - expenses grouped by subcategory
GET /api/v1/expenses?group_by=user - expenses grouped by user
GET /api/v1/expenses?group_by=category&order_dir=asc - categories ordered by total (lowest first)
GET /api/v1/expenses?group_by=category&user_id=1 - user 1
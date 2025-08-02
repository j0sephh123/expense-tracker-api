| Param           | Type        | Example    | Meaning                              |
| --------------- | ----------- | ---------- | ------------------------------------ |
| date_from       | string      | 2025-07-01 | Filter: start date                   |
| date_to         | string      | 2025-07-31 | Filter: end date                     |
| category_id     | int or list | 2 or 1,2,3 | One or more categories (OR logic)    |
| subcategory_id  | int or list | 4 or 4,5,6 | One or more subcategories (OR logic) |
| user_id         | int or list | 12 or 1,2  | Filter: user                         |
| order_by        | string      | amount     | Order: amount or date                |
| order_dir       | string      | desc       | Order: asc/desc                      |
| group_by        | string      | category   | Group: category/subcategory/user     |
| aggregates_only | boolean     | true       | Only group totals                    |

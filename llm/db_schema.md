```sql
CREATE TABLE categories (
  id int NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY name (name)
)
CREATE TABLE `expenses` (
  `id` int NOT NULL AUTO_INCREMENT,
  `amount` decimal(10,2) NOT NULL,
  `subcategory_id` int DEFAULT NULL,
  `user_id` int DEFAULT NULL,
  `note` text,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  KEY `expenses_ibfk_1` (`subcategory_id`),
  CONSTRAINT `expenses_ibfk_1` FOREIGN KEY (`subcategory_id`) REFERENCES `subcategories` (`id`) ON DELETE SET NULL,
  CONSTRAINT `expenses_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=1383 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE users (
  id int NOT NULL AUTO_INCREMENT,
  uid varchar(255) DEFAULT NULL,
  email varchar(255) NOT NULL,
  display_name varchar(255) DEFAULT NULL,
  created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY email (email),
  UNIQUE KEY uid (uid)
)
CREATE TABLE subcategories (
  id int NOT NULL AUTO_INCREMENT,
  category_id int NOT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY category_id (category_id,name),
  CONSTRAINT subcategories_ibfk_1 FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE RESTRICT
)
```
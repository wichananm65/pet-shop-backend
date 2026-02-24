BEGIN;

CREATE TEMP TABLE new_ids(pid int);

INSERT INTO products
    ("productName","productPrice","productDesc","productImg",
     "category","score","productNameTH","productDescTH","createdAt","updatedAt")
VALUES
  ('Happy Cat Food', 120, 'Premium dry food for cats', '/api/v1/product/0/image',
     'Cat snacks', 5, 'อาหารแมว','อาหารแมวคุณภาพสูง', now(), now()),
  ('Cat Snack Pack', 80, 'Crunchy treats your cat will love', '/api/v1/product/0/image',
     'Cat snacks', 3, 'ขนมแมว','ขนมแมวกรุบกรอบ', now(), now()),
  ('Interactive Cat Toy', 150, 'Battery-powered mouse toy for exercise', '/api/v1/product/0/image',
     'Cat exercise', 4, 'ของเล่นแมว','ของเล่นแมวเคลื่อนไหว', now(), now()),
  ('Self-cleaning Litter Box', 900, 'Easy-wash plastic litter box with cover', '/api/v1/product/0/image',
     'Sand and bathroom', 2, 'กระบะทราย','กระบะทรายทำความสะอาดง่าย', now(), now())
RETURNING "productID" INTO new_ids;

UPDATE products
SET "productImg" = '/api/v1/product/' || ni.pid || '/image'
FROM new_ids ni
WHERE products."productID" = ni.pid;

WITH added AS (
    SELECT p."productID" AS pid, p."category"
    FROM products p
    JOIN new_ids ni ON p."productID" = ni.pid
)
UPDATE category
SET "productID" = ARRAY[added.pid]
FROM added
WHERE category."categoryName" = added."category";

UPDATE products
SET product_img_data = pg_read_binary_file('/tmp/HappyCatFood.jpg')
WHERE "productID" = (SELECT pid FROM new_ids ORDER BY pid LIMIT 1);

UPDATE products
SET product_img_data = pg_read_binary_file('/tmp/CatSnack1.jpg')
WHERE "productID" = (SELECT pid FROM new_ids ORDER BY pid OFFSET 1 LIMIT 1);

UPDATE products
SET product_img_data = pg_read_binary_file('/tmp/CatToy1.jpg')
WHERE "productID" = (SELECT pid FROM new_ids ORDER BY pid OFFSET 2 LIMIT 1);

UPDATE products
SET product_img_data = pg_read_binary_file('/tmp/LitterBox.jpg')
WHERE "productID" = (SELECT pid FROM new_ids ORDER BY pid OFFSET 3 LIMIT 1);

COMMIT;

-- verification
SELECT "productID", "productImg", octet_length(product_img_data) AS bytes
FROM products
WHERE "productID" IN (SELECT pid FROM new_ids);

SELECT "categoryID", "categoryName", "productID"
FROM category
WHERE "productID" IS NOT NULL;

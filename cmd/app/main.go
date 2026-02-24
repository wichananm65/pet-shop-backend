package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/wichananm65/pet-shop-backend/internal/address"
	"github.com/wichananm65/pet-shop-backend/internal/banner"
	"github.com/wichananm65/pet-shop-backend/internal/cart"
	"github.com/wichananm65/pet-shop-backend/internal/category"
	"github.com/wichananm65/pet-shop-backend/internal/favorite"
	"github.com/wichananm65/pet-shop-backend/internal/order"
	"github.com/wichananm65/pet-shop-backend/internal/product"
	"github.com/wichananm65/pet-shop-backend/internal/recommended"
	shoppingmall "github.com/wichananm65/pet-shop-backend/internal/shopping-mall"
	"github.com/wichananm65/pet-shop-backend/internal/user"
)

func main() {
	_ = godotenv.Load()
	app := fiber.New()
	setupCORS(app)

	db := mustOpenDB()
	defer db.Close()

	// ensure bytea columns exist for storing images
	if _, err := db.Exec(`ALTER TABLE products ADD COLUMN IF NOT EXISTS product_img_data bytea, ADD COLUMN IF NOT EXISTS product_img_sec_data bytea`); err != nil {
		panic(err)
	}

	// ensure banner table exists; seed with public/banner images when empty
	// ensure user avatar column exists
	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_pic TEXT`); err != nil {
		panic(err)
	}
	// add mainAddressId to users if missing
	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS "mainAddressId" INT`); err != nil {
		panic(err)
	}
	// ensure cart product id array exists for storing user cart (legacy)
	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS "cartProductId" integer[]`); err != nil {
		panic(err)
	}
	// new JSONB map column for product->quantity. If existing column named cart is integer[]
	// convert it to jsonb preserving quantities (treat duplicate entries as qty 1).
	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS cart jsonb NOT NULL DEFAULT '{}'`); err != nil {
		panic(err)
	}
	// attempt conversion from integer[] to jsonb if the type differs
	if _, err := db.Exec(`ALTER TABLE users
    ALTER COLUMN cart TYPE jsonb
    USING to_jsonb(coalesce(cart, ARRAY[]::integer[]))`); err != nil {
		// ignore errors in case column already jsonb or conversion not applicable
	}
	// if some rows still hold a JSON array (e.g. [1,1,1]), collapse to map with counts
	if _, err := db.Exec(`UPDATE users
    SET cart = (
        SELECT jsonb_object_agg(elem::text, cnt)
        FROM (
            SELECT elem, count(*) AS cnt
            FROM unnest(cart::int[]) AS elem
            GROUP BY elem
        ) sub
    )
    WHERE jsonb_typeof(cart) = 'array'`); err != nil {
		// not fatal
		fmt.Printf("warning: cart normalization failed: %v\n", err)
	}

	// orders table storing cart map and price breakdown
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS orders (
        "orderID" SERIAL PRIMARY KEY,
        "userID" INT NOT NULL,
        cart jsonb NOT NULL DEFAULT '{}',
        quantity INT NOT NULL DEFAULT 0,
        "totalPrice" numeric NOT NULL DEFAULT 0,
        "shippingPrice" numeric NOT NULL DEFAULT 0,
        "grandPrice" numeric NOT NULL DEFAULT 0,
        status TEXT,
        "createdAt" TEXT,
        "updatedAt" TEXT
    )`); err != nil {
		panic(err)
	}
	// make sure the column exists in case table was created before cart added
	if _, err := db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS cart jsonb NOT NULL DEFAULT '{}'`); err != nil {
		panic(err)
	}
	// rename any lowercase columns from older schema versions
	if _, err := db.Exec(`ALTER TABLE orders RENAME COLUMN IF EXISTS totalprice TO "totalPrice"`); err != nil {
		// ignore
	}
	if _, err := db.Exec(`ALTER TABLE orders RENAME COLUMN IF EXISTS shippingprice TO "shippingPrice"`); err != nil {
		// ignore
	}
	if _, err := db.Exec(`ALTER TABLE orders RENAME COLUMN IF EXISTS grandprice TO "grandPrice"`); err != nil {
		// ignore
	}
	// if existing orders.cart is integer[] convert to jsonb
	if _, err := db.Exec(`ALTER TABLE orders
    ALTER COLUMN cart TYPE jsonb
    USING to_jsonb(coalesce(cart, ARRAY[]::integer[]))`); err != nil {
		// ignore
	}
	// normalize any array entries to map counts
	if _, err := db.Exec(`UPDATE orders
    SET cart = (
        SELECT jsonb_object_agg(elem::text, cnt)
        FROM (
            SELECT elem, count(*) AS cnt
            FROM unnest(cart::int[]) AS elem
            GROUP BY elem
        ) sub
    )
    WHERE jsonb_typeof(cart) = 'array'`); err != nil {
		fmt.Printf("warning: orders cart normalization failed: %v\n", err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS banner (banner_id SERIAL PRIMARY KEY, banner_img TEXT, banner_link TEXT, banner_alt TEXT, ord INT)`); err != nil {
		panic(err)
	}
	var bannerCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM banner`).Scan(&bannerCount); err == nil {
		if bannerCount == 0 {
			// seed a few banners that match files in frontend `public/banner`
			seed := []struct{ img, link, alt string }{
				{"/banner/04a429f3667447618ad41d1ddc3941295098953b.jpg", "", ""},
				{"/banner/3a2c4a01b382255d010fdce9b9c5942f82297af9.jpg", "", ""},
				{"/banner/8b1361654080c673a9ff07dd0f7ea6d51422c8b1 (1).jpg", "", ""},
			}
			for i, s := range seed {
				if _, err := db.Exec(`INSERT INTO banner (banner_img, banner_link, banner_alt, ord) VALUES ($1,$2,$3,$4)`, s.img, s.link, s.alt, len(seed)-i); err != nil {
					// ignore seed errors
					continue
				}
			}
		}
	}

	// ensure category table exists; seed with public/Category images when empty
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS category ("categoryID" SERIAL PRIMARY KEY, "categoryName" TEXT, "categoryNameTH" TEXT, "categoryImg" TEXT, ord INT)`); err != nil {
		panic(err)
	}
	// make sure new TH column exists if table pre‑dated change
	if _, err := db.Exec(`ALTER TABLE category ADD COLUMN IF NOT EXISTS "categoryNameTH" TEXT`); err != nil {
		panic(err)
	}

	// address table for storing user addresses (camelCase column names to match project convention)
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS address (
		"addressID" SERIAL PRIMARY KEY,
		"userID" INT NOT NULL,
		"addressDesc" TEXT,
		"phone" TEXT,
		"addressName" TEXT,
		"createdAt" TEXT,
		"updatedAt" TEXT
	)`); err != nil {
		panic(err)
	}
	// if previous installations used snake_case column names, rename them to camelCase
	if _, err := db.Exec(`DO $$
    BEGIN
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='address' AND column_name='address_id') THEN
            ALTER TABLE address RENAME COLUMN address_id TO "addressID";
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='address' AND column_name='user_id') THEN
            ALTER TABLE address RENAME COLUMN user_id TO "userID";
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='address' AND column_name='address_desc') THEN
            ALTER TABLE address RENAME COLUMN address_desc TO "addressDesc";
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='address' AND column_name='address_name') THEN
            ALTER TABLE address RENAME COLUMN address_name TO "addressName";
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='address' AND column_name='created_at') THEN
            ALTER TABLE address RENAME COLUMN created_at TO "createdAt";
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='address' AND column_name='updated_at') THEN
            ALTER TABLE address RENAME COLUMN updated_at TO "updatedAt";
        END IF;
    END
    $$;`); err != nil {
		// non-fatal - we can ignore failures here
		fmt.Printf("warning: could not rename address columns: %v\n", err)
	}
	// ensure userID column exists after possible renames; add if missing to avoid SQL errors
	if _, err := db.Exec(`ALTER TABLE address ADD COLUMN IF NOT EXISTS "userID" INT NOT NULL DEFAULT 0`); err != nil {
		fmt.Printf("warning: could not add missing userID column: %v\n", err)
	}
	var categoryCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM category`).Scan(&categoryCount); err == nil {
		if categoryCount == 0 {
			seed := []struct{ name, nameTH, img string }{
				{"Animal food", "อาหารสัตว์", "/Category/Animal _food.png"},
				{"Pet supplies", "ของใช้สัตว์เลี้ยง", "/Category/pet_supplies.png"},
				{"Clothes and accessories", "เสื้อผ้าและเครื่องแต่งกาย", "/Category/Clothes_and_accessories.png"},
				{"Cleaning equipment", "อุปกรณ์ทำความสะอาด", "/Category/Cleaning_equipment.png"},
				{"Sand and bathroom", "ทรายและห้องน้ำ", "/Category/sand_and_bathroom.png"},
				{"Hygiene care", "ปกป้องสุขภาพ", "/Category/Hygiene_care.png"},
				{"Cat snacks", "ขนมแมว", "/Category/Cat_snacks.png"},
				{"Cat exercise", "อุปกรณ์ออกกำลังกายแมว", "/Category/Cat_exercise.png"},
			}
			for i, s := range seed {
				if _, err := db.Exec(`INSERT INTO category ("categoryName", "categoryNameTH", "categoryImg", ord) VALUES ($1,$2,$3,$4)`, s.name, s.nameTH, s.img, len(seed)-i); err != nil {
					continue
				}
			}
		}
	}

	// ensure v2 product table exists and mirror any legacy data
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS product (
			product_id SERIAL PRIMARY KEY,
			product_name TEXT,
			product_name_en TEXT,
			category TEXT,
			product_price INT,
			score INT,
			product_desc TEXT,
			product_desc_en TEXT,
			product_pic TEXT,
			product_pic_second TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		);
	`); err != nil {
		fmt.Printf("warning: could not create product table: %v\n", err)
	}
	// copy any existing rows from legacy table into v2 if not already present
	if _, err := db.Exec(`
		INSERT INTO product (product_id, product_name, category, product_price, score, product_desc, product_pic, created_at, updated_at)
		SELECT "productID", "productName", category, "productPrice", score, "productDesc", "productImg", "createdAt", "updatedAt"
		FROM products p
		WHERE NOT EXISTS (SELECT 1 FROM product WHERE product_id = p."productID")
	`); err != nil {
		// non-fatal
	}

	// create user repo/service/handler so we can share the user service with the
	// new `favorite` handler (keeps favorite responsibilities isolated).
	userRepo := user.NewPostgresRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// build product service/handler early so we can reuse service elsewhere
	productRepo := product.NewPostgresRepository(db)
	productService := product.NewService(productRepo)
	productHandler := product.NewHandler(productService)

	jwtSecret := os.Getenv("JWT_SECRET")

	userHandler.RegisterPublicRoutes(app)

	// register recommended handler (internal/recommended)
	recommendedHandler := recommended.NewHandler(recommended.NewService(recommended.NewPostgresRepository(db)))
	recommendedHandler.RegisterPublicRoutes(app)

	// register banner handler (internal/banner)
	bannerHandler := banner.NewHandler(banner.NewService(banner.NewPostgresRepository(db)))
	bannerHandler.RegisterPublicRoutes(app)

	// register category handler (internal/category)
	categoryHandler := category.NewHandler(category.NewService(category.NewPostgresRepository(db)))
	categoryHandler.RegisterPublicRoutes(app)

	// register shopping-mall handler (internal/shopping-mall)
	shoppingMallHandler := shoppingmall.NewHandler(shoppingmall.NewService(shoppingmall.NewPostgresRepository(db)))
	shoppingMallHandler.RegisterPublicRoutes(app)

	// order handler (will register protected routes later)
	// it needs access to product service for enriching carts
	orderHandler := order.NewHandler(order.NewService(order.NewPostgresRepository(db)), userService, productService)

	// register product public routes after specific endpoints to avoid route param collision
	productHandler.RegisterPublicRoutes(app)

	// make uploaded files public
	app.Static("/uploads", "./uploads")

	// dev endpoint: import existing filesystem images into DB (public, gated by ALLOW_RESET_PRODUCTS)
	app.Post("/dev/import-product-images", func(c *fiber.Ctx) error {
		if os.Getenv("ALLOW_RESET_PRODUCTS") != "1" {
			return c.Status(fiber.StatusForbidden).SendString("not allowed")
		}
		rows, err := db.Query(`SELECT "productID", "productImg" FROM products`)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var id int
			var path sql.NullString
			if err := rows.Scan(&id, &path); err != nil {
				continue
			}
			if !path.Valid || path.String == "" || !strings.HasPrefix(path.String, "/uploads") {
				continue
			}
			p := "." + path.String
			b, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			if _, err := db.Exec(`UPDATE products SET product_img_data = $1 WHERE "productID" = $2`, b, id); err != nil {
				continue
			}
			count++
		}
		return c.JSON(fiber.Map{"imported": count})
	})

	// dev endpoint: drop, recreate and reseed `category` table (gated by ALLOW_RESET_CATEGORIES)
	app.Post("/dev/reset-categories", func(c *fiber.Ctx) error {
		if os.Getenv("ALLOW_RESET_CATEGORIES") != "1" {
			return c.Status(fiber.StatusForbidden).SendString("not allowed")
		}

		// drop and recreate table
		if _, err := db.Exec(`DROP TABLE IF EXISTS category`); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS category ("categoryID" SERIAL PRIMARY KEY, "categoryName" TEXT, "categoryNameTH" TEXT, "categoryImg" TEXT, ord INT)`); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		seed := []struct{ name, nameTH, img string }{
			{"Animal food", "อาหารสัตว์", "/Category/Animal _food.png"},
			{"Pet supplies", "ของใช้สัตว์เลี้ยง", "/Category/pet_supplies.png"},
			{"Clothes and accessories", "เสื้อผ้าและเครื่องแต่งกาย", "/Category/Clothes_and_accessories.png"},
			{"Cleaning equipment", "อุปกรณ์ทำความสะอาด", "/Category/Cleaning_equipment.png"},
			{"Sand and bathroom", "ทรายและห้องน้ำ", "/Category/sand_and_bathroom.png"},
			{"Hygiene care", "ปกป้องสุขภาพ", "/Category/Hygiene_care.png"},
			{"Cat snacks", "ขนมแมว", "/Category/Cat_snacks.png"},
			{"Cat exercise", "อุปกรณ์ออกกำลังกายแมว", "/Category/Cat_exercise.png"},
		}
		inserted := 0
		for i, s := range seed {
			if _, err := db.Exec(`INSERT INTO category ("categoryName", "categoryNameTH", "categoryImg", ord) VALUES ($1,$2,$3,$4)`, s.name, s.nameTH, s.img, len(seed)-i); err != nil {
				continue
			}
			inserted++
		}
		return c.JSON(fiber.Map{"inserted": inserted})
	})

	// public endpoint to serve product image bytes or fallback to file/redirect
	app.Get("/api/v1/product/:id<[0-9]+>/image", func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("invalid id")
		}

		var imgData []byte
		var path sql.NullString
		err = db.QueryRow(`SELECT product_img_data, "productImg" FROM products WHERE "productID" = $1`, id).Scan(&imgData, &path)
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("not found")
		}

		if len(imgData) > 0 {
			ct := http.DetectContentType(imgData)
			c.Set("Content-Type", ct)
			return c.Send(imgData)
		}

		if path.Valid && path.String != "" {
			if strings.HasPrefix(path.String, "/") {
				return c.SendFile("." + path.String)
			}
			return c.Redirect(path.String, fiber.StatusFound)
		}
		return c.Status(fiber.StatusNotFound).SendString("image not available")
	})

	app.Use(checkMiddleware)
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(jwtSecret),
		// allow unauthenticated GET requests for numeric product details and images
		Filter: func(c *fiber.Ctx) bool {
			// allow unauthenticated GETs for numeric product id paths
			p := c.Path()
			fmt.Printf("[DEBUG] jwt.Filter invoked — method=%s path=%s\n", c.Method(), p)
			if c.Method() != "GET" {
				fmt.Printf("[DEBUG] jwt.Filter -> not GET, require auth\n")
				return false
			}
			if strings.HasPrefix(p, "/api/v1/product/") {
				rest := strings.TrimPrefix(p, "/api/v1/product/")
				seg := strings.SplitN(rest, "/", 2)[0]
				if _, err := strconv.Atoi(seg); err == nil {
					fmt.Printf("[DEBUG] jwt.Filter -> public product GET, skipping JWT: id=%s\n", seg)
					return true // skip JWT check for public product GETs
				}
			}
			fmt.Printf("[DEBUG] jwt.Filter -> require auth\n")
			return false
		},
	}))

	userHandler.RegisterProtectedRoutes(app)
	// order endpoints (protected)
	orderHandler.RegisterProtectedRoutes(app)
	// favorites are handled by a dedicated handler with its own repository/service
	favoriteRepo := favorite.NewPostgresRepository(db)
	favoriteService := favorite.NewService(favoriteRepo)
	favoriteHandler := favorite.NewHandler(favoriteService)
	favoriteHandler.RegisterProtectedRoutes(app)

	// address endpoints
	addressRepo := address.NewPostgresRepository(db)
	addressService := address.NewService(addressRepo)
	addressHandler := address.NewHandler(addressService)
	addressHandler.RegisterProtectedRoutes(app)

	// cart endpoints
	cartRepo := cart.NewPostgresRepository(db)
	cartService := cart.NewService(cartRepo)
	cartHandler := cart.NewHandler(cartService)
	cartHandler.RegisterProtectedRoutes(app)

	productHandler.RegisterProtectedRoutes(app)

	// protected endpoint to upload and persist image bytes into Postgres
	app.Post("/api/v1/product/:id<[0-9]+>/image", func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("invalid id")
		}

		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("file is required")
		}
		f, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		if _, err := db.Exec(`UPDATE products SET product_img_data = $1 WHERE "productID" = $2`, b, id); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		return c.SendString("ok")
	})

	// upload endpoint remains protected
	app.Post("/upload", uploadFile)

	app.Listen(":8080")
}

func setupCORS(app *fiber.App) {
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
}

func mustOpenDB() *sql.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		panic("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return db
}

func uploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if err := c.SaveFile(file, "./uploads/"+file.Filename); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendString("File uploaded successfully: " + file.Filename)
}

func checkMiddleware(c *fiber.Ctx) error {
	start := time.Now()
	fmt.Printf("URL = %s, Method = %s, Start Time = %v\n", c.OriginalURL(), c.Method(), start)
	return c.Next()
}

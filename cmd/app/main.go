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
	"github.com/wichananm65/pet-shop-backend/internal/banner"
	"github.com/wichananm65/pet-shop-backend/internal/product"
	"github.com/wichananm65/pet-shop-backend/internal/recommended"
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

	userHandler := buildUserHandler(db)
	productHandler := buildProductHandler(db)
	jwtSecret := os.Getenv("JWT_SECRET")

	userHandler.RegisterPublicRoutes(app)
	productHandler.RegisterPublicRoutes(app)

	// register recommended handler (internal/recommended)
	recommendedHandler := recommended.NewHandler(recommended.NewService(recommended.NewPostgresRepository(db)))
	recommendedHandler.RegisterPublicRoutes(app)

	// register banner handler (internal/banner)
	bannerHandler := banner.NewHandler(banner.NewService(banner.NewPostgresRepository(db)))
	bannerHandler.RegisterPublicRoutes(app)

	// Public API for shopping mall section — returns only id + image
	app.Get("/api/v1/product/shopping-mall", func(c *fiber.Ctx) error {
		rows, err := db.Query(`SELECT "productID" FROM products ORDER BY "productID" LIMIT 100`)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		defer rows.Close()

		type Item struct {
			ProductId  int     `json:"productId"`
			ProductPic *string `json:"productPic,omitempty"`
		}

		out := make([]Item, 0)
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				continue
			}
			// return API endpoint for the image — frontend will call this URL
			u := fmt.Sprintf("/api/v1/product/%d/image", id)
			out = append(out, Item{ProductId: id, ProductPic: &u})
		}
		return c.JSON(out)
	})

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

	// public endpoint to serve product image bytes or fallback to file/redirect
	app.Get("/api/v1/product/:id/image", func(c *fiber.Ctx) error {
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
	}))

	userHandler.RegisterProtectedRoutes(app)
	productHandler.RegisterProtectedRoutes(app)

	// protected endpoint to upload and persist image bytes into Postgres
	app.Post("/api/v1/product/:id/image", func(c *fiber.Ctx) error {
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

func buildUserHandler(db *sql.DB) *user.Handler {
	userRepo := user.NewPostgresRepository(db)
	userService := user.NewService(userRepo)
	return user.NewHandler(userService)
}

func buildProductHandler(db *sql.DB) *product.Handler {
	productRepo := product.NewPostgresRepository(db)
	productService := product.NewService(productRepo)
	return product.NewHandler(productService)
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

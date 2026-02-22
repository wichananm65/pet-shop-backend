package user

import (
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// helper to build an app with a simple "bootstrap" middleware that injects a
// jwt.Token into locals when the X-User-ID header is provided. This avoids
// pulling in the full jwtware middleware and keeps tests lightweight.
func makeAppWithUserHandler(uHandler *Handler) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		if v := c.Get("X-User-ID"); v != "" {
			id, err := strconv.Atoi(v)
			if err == nil {
				claims := jwt.MapClaims{"user_id": id}
				tok := &jwt.Token{Claims: claims}
				c.Locals("user", tok)
			}
		}
		return c.Next()
	})
	uHandler.RegisterProtectedRoutes(app)
	return app
}

func TestProfileRoute_RegistrationAndAuth(t *testing.T) {
	seed := []User{{ID: 7, Email: "j@example.com", FirstName: "Jenny", LastName: "Test", Phone: "123", Gender: "F", MainAddressID: func() *int { i := 99; return &i }()}}
	repo := NewInMemoryRepository(seed)
	service := NewService(repo)
	handler := NewHandler(service)
	app := makeAppWithUserHandler(handler)

	// route registration check
	routes := map[string]bool{}
	for _, grp := range app.Stack() {
		for _, r := range grp {
			routes[r.Path] = true
		}
	}
	if !routes["/api/v1/profile"] {
		t.Fatalf("expected route '/api/v1/profile' to be registered")
	}

	// unauthorized request should yield 401
	req := httptest.NewRequest("GET", "/api/v1/profile", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("profile request failed: %v", err)
	}
	if res.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected unauthorized status, got %d", res.StatusCode)
	}

	// authorized request using X-User-ID header
	req2 := httptest.NewRequest("GET", "/api/v1/profile", nil)
	req2.Header.Set("X-User-ID", "7")
	res2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("authorized profile request failed: %v", err)
	}
	if res2.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK for authorized profile, got %d", res2.StatusCode)
	}

	// read body and ensure returned user matches and password is blank
	b, _ := io.ReadAll(res2.Body)
	body := string(b)
	if !strings.Contains(body, "j@example.com") {
		t.Fatalf("response body does not contain expected email, got %s", body)
	}
	if !strings.Contains(body, "mainAddressId") {
		t.Fatalf("response body does not include mainAddressId, got %s", body)
	}
	if strings.Contains(body, "password") {
		t.Fatalf("response body should not expose password field")
	}
}

func TestProfileUpdateAndAvatar(t *testing.T) {
	seed := []User{{ID: 15, Email: "u15@example.com", FirstName: "Old", LastName: "Name", Phone: "000", Gender: "male"}}
	repo := NewInMemoryRepository(seed)
	service := NewService(repo)
	handler := NewHandler(service)
	app := makeAppWithUserHandler(handler)

	// update profile fields using both PUT and PATCH to ensure both
	// methods are accepted by the handler.
	updateJSON := `{"firstName":"New","lastName":"User","phone":"999","gender":"female"}`

	for _, method := range []string{"PUT", "PATCH"} {
		req := httptest.NewRequest(method, "/api/v1/profile", strings.NewReader(updateJSON))
		req.Header.Set("X-User-ID", "15")
		req.Header.Set("Content-Type", "application/json")
		res, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s update request failed: %v", method, err)
		}
		if res.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200 OK on %s update, got %d", method, res.StatusCode)
		}
		b, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(b), "New") {
			t.Fatalf("updated response missing new name for %s: %s", method, string(b))
		}
	}
	// setting mainAddressId should be supported
	mainJSON := `{"mainAddressId":42}`
	reqMain := httptest.NewRequest("PATCH", "/api/v1/profile", strings.NewReader(mainJSON))
	reqMain.Header.Set("X-User-ID", "15")
	reqMain.Header.Set("Content-Type", "application/json")
	resMain, err := app.Test(reqMain)
	if err != nil {
		t.Fatalf("main address update failed: %v", err)
	}
	if resMain.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on main address update, got %d", resMain.StatusCode)
	}
	bMain, _ := io.ReadAll(resMain.Body)
	if !strings.Contains(string(bMain), "mainAddressId") {
		t.Fatalf("response missing mainAddressId: %s", string(bMain))
	}
	uFinal, _ := repo.GetByID(15)
	if uFinal.MainAddressID == nil || *uFinal.MainAddressID != 42 {
		t.Fatalf("mainAddressId not persisted: %+v", uFinal)
	}

	// avatar upload: create a fake file by POSTing to the dedicated
	// endpoint (existing behaviour) to ensure that still works.
	body := &strings.Builder{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "avatar.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte("PNGDATA"))
	writer.Close()

	req2 := httptest.NewRequest("POST", "/api/v1/profile/avatar", strings.NewReader(body.String()))
	req2.Header.Set("X-User-ID", "15")
	req2.Header.Set("Content-Type", writer.FormDataContentType())
	res2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("avatar upload failed: %v", err)
	}
	if res2.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK avatar upload, got %d", res2.StatusCode)
	}
	b2, _ := io.ReadAll(res2.Body)
	if !strings.Contains(string(b2), "avatarPic") {
		t.Fatalf("response should include avatarPic, got %s", string(b2))
	}
	// verify in-memory repo user has avatar path prefixed
	u, _ := repo.GetByID(15)
	if u.AvatarPic == nil || *u.AvatarPic == "" {
		t.Fatalf("avatar pic not set in repo")
	}

	// now exercise the new combined update route twice: once using the generic
	// "file" key and once using the preferred "avatarPic" key, verifying both
	// are accepted.
	// first, verify both field names still work for uploads
	for _, key := range []string{"file", "avatarPic"} {
		body3 := &strings.Builder{}
		writer3 := multipart.NewWriter(body3)
		writer3.WriteField("firstName", "Combined")
		writer3.WriteField("lastName", "Request")
		writer3.WriteField("phone", "222")
		writer3.WriteField("gender", "other")
		part3, err := writer3.CreateFormFile(key, "avatar2.png")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		part3.Write([]byte("PNGDATA2"))
		writer3.Close()

		req3 := httptest.NewRequest("PUT", "/api/v1/profile", strings.NewReader(body3.String()))
		req3.Header.Set("X-User-ID", "15")
		req3.Header.Set("Content-Type", writer3.FormDataContentType())
		res3, err := app.Test(req3)
		if err != nil {
			t.Fatalf("combined update failed using %s: %v", key, err)
		}
		if res3.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200 OK on combined update using %s, got %d", key, res3.StatusCode)
		}
		b3, _ := io.ReadAll(res3.Body)
		if !strings.Contains(string(b3), "avatarPic") || !strings.Contains(string(b3), "Combined") {
			t.Fatalf("combined response missing expected data using %s: %s", key, string(b3))
		}
		// confirm repo user updated
		u2, _ := repo.GetByID(15)
		if u2.FirstName != "Combined" || u2.AvatarPic == nil {
			t.Fatalf("combined update did not persist changes using %s: %+v", key, u2)
		}
	}
	// also test remove flag when no file is supplied
	body4 := &strings.Builder{}
	writer4 := multipart.NewWriter(body4)
	writer4.WriteField("firstName", "Delete")
	writer4.WriteField("removeAvatar", "true")
	writer4.Close()

	req4 := httptest.NewRequest("PUT", "/api/v1/profile", strings.NewReader(body4.String()))
	req4.Header.Set("X-User-ID", "15")
	req4.Header.Set("Content-Type", writer4.FormDataContentType())
	res4, err := app.Test(req4)
	if err != nil {
		t.Fatalf("remove update failed: %v", err)
	}
	if res4.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on remove update, got %d", res4.StatusCode)
	}
	b4, _ := io.ReadAll(res4.Body)
	if !strings.Contains(string(b4), "avatarPic") {
		t.Fatalf("remove response should include avatarPic key: %s", string(b4))
	}
	u3, _ := repo.GetByID(15)
	if u3.AvatarPic != nil {
		t.Fatalf("avatar not cleared on remove update: %+v", u3)
	}

	// now exercise removal flag: set removeAvatar=true and ensure avatar cleared
	body5 := &strings.Builder{}
	writer5 := multipart.NewWriter(body5)
	writer5.WriteField("firstName", "Removed")
	writer5.WriteField("removeAvatar", "true")
	writer5.Close()

	req5 := httptest.NewRequest("PUT", "/api/v1/profile", strings.NewReader(body5.String()))
	req5.Header.Set("X-User-ID", "15")
	req5.Header.Set("Content-Type", writer5.FormDataContentType())
	res5, err := app.Test(req5)
	if err != nil {
		t.Fatalf("avatar removal update failed: %v", err)
	}
	if res5.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on avatar removal update, got %d", res5.StatusCode)
	}
	b5, _ := io.ReadAll(res5.Body)
	if !strings.Contains(string(b5), "Removed") {
		t.Fatalf("removal response missing updated name: %s", string(b5))
	}
	u4, _ := repo.GetByID(15)
	if u4.FirstName != "Removed" || u4.AvatarPic != nil {
		t.Fatalf("avatar was not cleared by removal request: %+v", u4)
	}
}

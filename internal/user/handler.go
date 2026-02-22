package user

import (
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type Handler struct {
	service *Service
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Gender    string `json:"gender"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterPublicRoutes(app *fiber.App) {
	app.Post("/api/v1/sign-in", h.login)
	app.Post("/api/v1/sign-up", h.register)
}

func (h *Handler) RegisterProtectedRoutes(app *fiber.App) {
	app.Get("/users", h.getUsers)
	app.Get("/user/:id", h.getUser)
	app.Post("/users", h.createUser)
	app.Put("/user/:id", h.updateUser)
	app.Delete("/user/:id", h.deleteUser)
	// profile endpoint returns the current user based on JWT claims
	app.Get("/api/v1/profile", h.getProfile)
	// support both PUT and PATCH for updating profile fields. the handler
	// accepts partial payloads so PATCH behaviour is satisfied.
	app.Put("/api/v1/profile", h.updateProfile)
	app.Patch("/api/v1/profile", h.updateProfile)
	app.Post("/api/v1/profile/avatar", h.uploadAvatar)
	app.Delete("/api/v1/profile/avatar", h.removeAvatar)

}

func (h *Handler) login(c *fiber.Ctx) error {
	payload := new(loginRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	user, err := h.service.Authenticate(payload.Email, payload.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid email or password"})
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"user":    sanitizeUser(user),
		"token":   signed,
	})
}

func (h *Handler) register(c *fiber.Ctx) error {
	payload := new(registerRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if payload.isMissingRequiredFields() {
		return c.Status(fiber.StatusBadRequest).SendString("Missing required fields")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	created, err := h.service.Register(User{
		Email:     payload.Email,
		Password:  payload.Password,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Phone:     payload.Phone,
		Gender:    payload.Gender,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		if err == ErrEmailExists {
			return c.Status(fiber.StatusConflict).SendString("Email already exists")
		}
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(sanitizeUser(created))
}

func (h *Handler) getUsers(c *fiber.Ctx) error {
	users := h.service.List()
	response := make([]User, 0, len(users))
	for _, user := range users {
		response = append(response, sanitizeUser(user))
	}
	return c.JSON(response)
}

func (h *Handler) getUser(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	user, err := h.service.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	return c.JSON(sanitizeUser(user))
}

// getProfile returns the user record for the currently authenticated user.
// It reads the user_id claim from the JWT and then loads the user from the
// service. The returned object is sanitized so the password field is blank.
func (h *Handler) getProfile(c *fiber.Ctx) error {
	userID, err := GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	user, err := h.service.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
	}

	return c.JSON(sanitizeUser(user))
}

// profileUpdateRequest represents the fields the client may send to update.
type profileUpdateRequest struct {
	FirstName     *string `json:"firstName,omitempty"`
	LastName      *string `json:"lastName,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Gender        *string `json:"gender,omitempty"`
	MainAddressID *int    `json:"mainAddressId,omitempty"`
	RemoveAvatar  *string `json:"removeAvatar,omitempty"`
}

func (h *Handler) updateProfile(c *fiber.Ctx) error {
	userID, err := GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	var rmFlag string
	existing, err := h.service.GetByID(userID)
	// detect multipart via header prefix rather than c.Is which can misbehave
	ct := c.Get("Content-Type")
	isMultipart := strings.HasPrefix(ct, "multipart/form-data")
	// if this is a multipart request, parse it so FormValue works reliably
	if isMultipart {
		c.MultipartForm() // ignore error
	}
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
	}

	// support both JSON and multipart requests
	if isMultipart {
		// pull values from the form directly
		if v := c.FormValue("firstName"); v != "" {
			existing.FirstName = v
		}
		if v := c.FormValue("lastName"); v != "" {
			existing.LastName = v
		}
		if v := c.FormValue("phone"); v != "" {
			existing.Phone = v
		}
		if v := c.FormValue("gender"); v != "" {
			existing.Gender = v
		}
		if v := c.FormValue("mainAddressId"); v != "" {
			if id, err := strconv.Atoi(v); err == nil {
				existing.MainAddressID = &id
			}
		}
		rmFlag = c.FormValue("removeAvatar")
	} else {
		// normal JSON body
		var payload profileUpdateRequest
		c.BodyParser(&payload) // ignore error
		if payload.FirstName != nil {
			existing.FirstName = *payload.FirstName
		}
		if payload.LastName != nil {
			existing.LastName = *payload.LastName
		}
		if payload.Phone != nil {
			existing.Phone = *payload.Phone
		}
		if payload.Gender != nil {
			existing.Gender = *payload.Gender
		}
		if payload.MainAddressID != nil {
			existing.MainAddressID = payload.MainAddressID
		}
		if payload.RemoveAvatar != nil && *payload.RemoveAvatar == "true" {
			rmFlag = "true"
		}
	}

	// client may signal that the avatar should be removed rather than
	// uploaded. this takes precedence over any file that might be present.
	if rmFlag != "" {
		existing.AvatarPic = nil
	} else {
		// if the client sent a file (multipart request) then treat it as an avatar
		// upload and update avatar path at the same time. support both the
		// descriptive "avatarPic" field name and the generic "file" key.
		var file *multipart.FileHeader
		if f, ferr := c.FormFile("avatarPic"); ferr == nil && f != nil {
			file = f
		} else if f, ferr := c.FormFile("file"); ferr == nil && f != nil {
			file = f
		}
		if file != nil {
			f, err := file.Open()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
			}
			defer f.Close()
			path := "/uploads/avatars/" + strconv.Itoa(userID) + "_" + file.Filename
			dest := "." + path
			if err := os.MkdirAll("./uploads/avatars", 0755); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
			}
			if err := c.SaveFile(file, dest); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
			}
			existing.AvatarPic = &path
		}
	}

	updated, err := h.service.Update(userID, existing)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	// build explicit response so avatarPic is always present (even if nil)
	resp := sanitizeUser(updated)
	return c.JSON(fiber.Map{"avatarPic": resp.AvatarPic, "user": resp})
}

func (h *Handler) uploadAvatar(c *fiber.Ctx) error {
	userID, err := GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	// clients may send the avatar as either the generic "file" key or the
	// more descriptive "avatarPic" key depending on implementation. prefer
	// the latter if present for clarity.
	var file *multipart.FileHeader
	if f, e := c.FormFile("avatarPic"); e == nil && f != nil {
		file = f
	} else if f, e := c.FormFile("file"); e == nil && f != nil {
		file = f
	}
	if file == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "file is required"})
	}
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	defer f.Close()

	// save under uploads/avatars/userID_filename
	path := "/uploads/avatars/" + strconv.Itoa(userID) + "_" + file.Filename
	dest := "." + path
	// ensure directory exists
	if err := os.MkdirAll("./uploads/avatars", 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	if err := c.SaveFile(file, dest); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	// update user record
	existing, err := h.service.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
	}
	existing.AvatarPic = &path
	updated, err := h.service.Update(userID, existing)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	return c.JSON(fiber.Map{"avatarPic": path, "user": sanitizeUser(updated)})
}

func (h *Handler) removeAvatar(c *fiber.Ctx) error {
	userID, err := GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	existing, err := h.service.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
	}
	existing.AvatarPic = nil
	updated, err := h.service.Update(userID, existing)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	return c.JSON(fiber.Map{"avatarPic": nil, "user": sanitizeUser(updated)})
}

func (h *Handler) createUser(c *fiber.Ctx) error {
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	created, err := h.service.Create(*user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(sanitizeUser(created))
}

func (h *Handler) updateUser(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	userUpdate := new(User)
	if err := c.BodyParser(userUpdate); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	updated, err := h.service.Update(userID, *userUpdate)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	return c.JSON(sanitizeUser(updated))
}

func (h *Handler) deleteUser(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if err := h.service.Delete(userID); err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	return c.SendString("User deleted")
}

func (r registerRequest) isMissingRequiredFields() bool {
	return r.Email == "" || r.Password == "" || r.FirstName == "" || r.LastName == "" || r.Phone == "" || r.Gender == ""
}

// GetUserIDFromCtx extracts the user_id claim from the JWT token stored
// in `c.Locals("user")`. This duplicate logic used by several packages,
// so we export it here for reuse.
func GetUserIDFromCtx(c *fiber.Ctx) (int, error) {
	u := c.Locals("user")
	if u == nil {
		return 0, fiber.ErrUnauthorized
	}
	tok, ok := u.(*jwt.Token)
	if !ok {
		return 0, fiber.ErrUnauthorized
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fiber.ErrUnauthorized
	}
	if raw, ok := claims["user_id"]; ok {
		switch v := raw.(type) {
		case float64:
			return int(v), nil
		case int:
			return v, nil
		case int64:
			return int(v), nil
		case string:
			id, err := strconv.Atoi(v)
			if err != nil {
				return 0, fiber.ErrUnauthorized
			}
			return id, nil
		default:
			return 0, fiber.ErrUnauthorized
		}
	}
	return 0, fiber.ErrUnauthorized
}

func sanitizeUser(user User) User {
	user.Password = ""
	return user
}

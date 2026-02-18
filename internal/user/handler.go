package user

import (
	"os"
	"strconv"
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

func sanitizeUser(user User) User {
	user.Password = ""
	return user
}

package presenter

import "pet-shop-backend/internal/domain/entity"

// UserPresenter shapes domain entities for delivery layer responses.
type UserPresenter struct{}

func NewUserPresenter() *UserPresenter {
	return &UserPresenter{}
}

type UserResponse struct {
	ID        int64  `json:"user_id"`
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Phone     string `json:"phone,omitempty"`
	Gender    string `json:"gender,omitempty"`
	AvatarPic string `json:"avatar_pic,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (p *UserPresenter) ToResponse(user *entity.User) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Phone:     user.Phone,
		Gender:    user.Gender,
		AvatarPic: user.AvatarPic,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (p *UserPresenter) ToList(users []*entity.User) []*UserResponse {
	result := make([]*UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, p.ToResponse(user))
	}
	return result
}

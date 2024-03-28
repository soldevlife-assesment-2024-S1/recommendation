package response

type UserServiceValidate struct {
	IsValid bool `json:"is_valid"`
	UserID  int  `json:"user_id"`
}

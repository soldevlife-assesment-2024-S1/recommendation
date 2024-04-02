package response

type UserServiceValidate struct {
	IsValid bool `json:"is_valid"`
	UserID  int  `json:"user_id"`
}

type Recomendation struct {
	RegionName         string  `json:"region_name"`
	Level              string  `json:"level"`
	DiscountPrice      int     `json:"discount_price"`
	OriginalPrice      int     `json:"original_price"`
	DiscountPercentage float64 `json:"discount_percentage"`
}

type UserProfile struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Region   string `json:"region"`
}

type Ticket struct {
	RegionName string `json:"region_name"`
	Level      string `json:"level"`
	Price      int    `json:"price"`
}

type BreTicketDiscounted struct {
	DiscountAmount  int `json:"discount_amount"`
	DiscountedPrice int `json:"discounted_price"`
}

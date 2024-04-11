package response

type BaseResponse struct {
	Meta interface{} `json:"meta"`
	Data interface{} `json:"data"`
}

type UserServiceValidate struct {
	IsValid   bool   `json:"is_valid"`
	UserID    int64  `json:"user_id"`
	EmailUser string `json:"email_user"`
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

type OnlineTicket struct {
	IsSoldOut      bool `json:"is_sold_out"`
	IsFirstSoldOut bool `json:"is_first_sold_out"`
}

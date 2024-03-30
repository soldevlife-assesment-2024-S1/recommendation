package request

type Pagination struct {
	Page int `json:"page" form:"page" required:"true" validate:"required,numeric"`
	Size int `json:"size" form:"size" required:"true" validate:"required,numeric"`
}

type PoisonedQueue struct {
	TopicTarget string      `json:"topic_target" validate:"required"`
	ErrorMsg    string      `json:"error_msg" validate:"required"`
	Payload     interface{} `json:"payload" validate:"required"`
}

type UpdateVenueStatus struct {
	VenueName      string `json:"venue_name" validate:"required"`
	IsSoldOut      bool   `json:"is_sold_out" validate:"required"`
	IsFirstSoldOut bool   `json:"is_first_sold_out" validate:"required"`
}

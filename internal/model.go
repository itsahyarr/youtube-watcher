package internal

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScrapeLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	URL        string             `bson:"url" json:"url"`
	Target     string             `bson:"target" json:"target"`
	Action     string             `bson:"action" json:"action"`
	Status     string             `bson:"status" json:"status"`
	HTTPStatus int                `bson:"httpStatus" json:"httpStatus"`
	Message    string             `bson:"message" json:"message"`
	Error      *string            `bson:"error,omitempty" json:"error,omitempty"`
	StartedAt  time.Time          `bson:"startedAt" json:"startedAt"`
	FinishedAt time.Time          `bson:"finishedAt" json:"finishedAt"`
	DurationMs int64              `bson:"durationMs" json:"durationMs"`
	ProxyUsed  *string            `bson:"proxyUsed,omitempty" json:"proxyUsed,omitempty"`
	ExitIP     *string            `bson:"exitIp,omitempty" json:"exitIp,omitempty"`
	Headless   bool               `bson:"headless" json:"headless"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
}

type ScrapeData struct {
	LogID   string `json:"logId" example:"64f000000000000000000000"`
	URL     string `json:"url" example:"https://www.youtube.com/watch?v=abc123"`
	Action  string `json:"action" example:"CLICK_PLAY"`
	Result  string `json:"result" example:"SUCCESS"`
	Message string `json:"message" example:"YouTube video play button clicked successfully"`
}

type SuccessResponse struct {
	Code    int        `json:"code" example:"200"`
	Status  string     `json:"status" example:"OK"`
	Success bool       `json:"success" example:"true"`
	Data    ScrapeData `json:"data"`
}

type ErrorResponse struct {
	Code    int                    `json:"code" example:"400"`
	Status  string                 `json:"status" example:"BAD_REQUEST"`
	Success bool                   `json:"success" example:"false"`
	Errors  map[string]interface{} `json:"errors"`
}

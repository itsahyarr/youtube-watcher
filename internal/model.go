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
	Headless   bool               `bson:"headless" json:"headless"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
}

type SuccessResponse struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Code    int                    `json:"code"`
	Status  string                 `json:"status"`
	Success bool                   `json:"success"`
	Errors  map[string]interface{} `json:"errors"`
}

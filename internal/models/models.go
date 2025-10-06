package models

import "time"

type Url struct {
	LongUrl     string    `json:"long_url"`
	Slug        string    `json:"slug"`
	CustomAlias string    `json:"alias"`
	Created_at  time.Time `json:"created_at"`
	Expires_at  time.Time `json:"expires_at"`
}

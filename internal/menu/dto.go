package menu

import "time"

type menuRequest struct {
	Title      string        `json:"title"`
	Note       string        `json:"note"`
	DateKey    string        `json:"date_key"`
	Time       string        `json:"time"`
	Status     string        `json:"status"`
	DinerCount int           `json:"diner_count"`
	RecipeIDs  []string      `json:"recipe_ids"`
	Dishes     []dishRequest `json:"dishes"`
}

type menuStatusRequest struct {
	Status string `json:"status"`
}

type dishRequest struct {
	RecipeID string `json:"recipe_id"`
	Name     string `json:"name"`
	Count    int    `json:"count"`
}

type menuResponse struct {
	ID         string         `json:"id"`
	Title      string         `json:"title"`
	Note       string         `json:"note"`
	DateKey    string         `json:"date_key"`
	DateLabel  string         `json:"date_label"`
	Weekday    string         `json:"weekday"`
	Time       string         `json:"time"`
	Status     string         `json:"status"`
	DinerCount int            `json:"diner_count"`
	RecipeIDs  []string       `json:"recipe_ids"`
	Dishes     []dishResponse `json:"dishes"`
	Tone       string         `json:"tone"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type dishResponse struct {
	RecipeID string `json:"recipe_id"`
	Name     string `json:"name"`
	Count    int    `json:"count"`
}

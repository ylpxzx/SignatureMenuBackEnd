package recipe

import "time"

type recipeRequest struct {
	Name             string              `json:"name"`
	Description      string              `json:"description"`
	ServingCount     int                 `json:"serving_count"`
	EstimatedMinutes int                 `json:"estimated_minutes"`
	Difficulty       int                 `json:"difficulty"`
	IsAvailable      *bool               `json:"is_available"`
	Ingredients      []ingredientRequest `json:"ingredients"`
	Steps            []stepRequest       `json:"steps"`
}

type ingredientRequest struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
	Note   string `json:"note"`
}

type stepRequest struct {
	ID               string `json:"id"`
	StepOrder        int    `json:"step_order"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	EstimatedMinutes int    `json:"estimated_minutes"`
}

type recipeResponse struct {
	ID               string               `json:"id"`
	Name             string               `json:"name"`
	Description      string               `json:"description"`
	ServingCount     int                  `json:"serving_count"`
	EstimatedMinutes int                  `json:"estimated_minutes"`
	Difficulty       int                  `json:"difficulty"`
	IsAvailable      bool                 `json:"is_available"`
	Ingredients      []ingredientResponse `json:"ingredients"`
	Steps            []stepResponse       `json:"steps"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

type ingredientResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Amount    string    `json:"amount"`
	Unit      string    `json:"unit"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type stepResponse struct {
	ID               string    `json:"id"`
	StepOrder        int       `json:"step_order"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	EstimatedMinutes int       `json:"estimated_minutes"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

package store

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"display_name"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Recipe struct {
	ID               string       `json:"id"`
	UserID           string       `json:"user_id"`
	Name             string       `json:"name"`
	Description      string       `json:"description"`
	ServingCount     int          `json:"serving_count"`
	EstimatedMinutes int          `json:"estimated_minutes"`
	Difficulty       int          `json:"difficulty"`
	IsAvailable      bool         `json:"is_available"`
	Ingredients      []Ingredient `json:"ingredients"`
	Steps            []RecipeStep `json:"steps"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
	DeletedAt        *time.Time   `json:"deleted_at,omitempty"`
}

type Ingredient struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Amount    string    `json:"amount"`
	Unit      string    `json:"unit"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RecipeStep struct {
	ID               string    `json:"id"`
	StepOrder        int       `json:"step_order"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	EstimatedMinutes int       `json:"estimated_minutes"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type RecipeMutation struct {
	Name             string
	Description      string
	ServingCount     int
	EstimatedMinutes int
	Difficulty       int
	IsAvailable      bool
	Ingredients      []IngredientMutation
	Steps            []StepMutation
}

type IngredientMutation struct {
	ID     string
	Name   string
	Amount string
	Unit   string
	Note   string
}

type StepMutation struct {
	ID               string
	StepOrder        int
	Title            string
	Description      string
	EstimatedMinutes int
}

type IngredientSummary struct {
	Name        string `json:"name"`
	RecipeCount int    `json:"recipe_count"`
}

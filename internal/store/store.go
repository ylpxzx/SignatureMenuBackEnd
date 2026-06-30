package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Store struct {
	mu   sync.RWMutex
	path string
	data appData
}

type appData struct {
	Users   []User   `json:"users"`
	Recipes []Recipe `json:"recipes"`
}

func New(path string) (*Store, error) {
	store := &Store{
		path: path,
		data: appData{
			Users:   []User{},
			Recipes: []Recipe{},
		},
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return store, store.saveLocked()
		}
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&store.data); err != nil {
		return nil, err
	}
	if store.data.Users == nil {
		store.data.Users = []User{}
	}
	if store.data.Recipes == nil {
		store.data.Recipes = []Recipe{}
	}

	return store, nil
}

func (s *Store) CreateUser(username string, passwordHash string, displayName string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	username = normalizeUsername(username)
	if username == "" || passwordHash == "" {
		return User{}, ErrInvalidInput
	}
	if displayName = strings.TrimSpace(displayName); displayName == "" {
		displayName = username
	}

	for _, user := range s.data.Users {
		if user.Username == username {
			return User{}, ErrConflict
		}
	}

	now := time.Now().UTC()
	user := User{
		ID:           newID(),
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.data.Users = append(s.data.Users, user)
	if err := s.saveLocked(); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Store) GetUser(id string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.data.Users {
		if user.ID == id {
			return user, nil
		}
	}
	return User{}, ErrNotFound
}

func (s *Store) FindUserByUsername(username string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	username = normalizeUsername(username)
	for _, user := range s.data.Users {
		if user.Username == username {
			return user, nil
		}
	}
	return User{}, ErrNotFound
}

func (s *Store) ListRecipes(userID string) []Recipe {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recipes := make([]Recipe, 0)
	for _, recipe := range s.data.Recipes {
		if recipe.UserID == userID && recipe.DeletedAt == nil {
			recipes = append(recipes, recipe)
		}
	}

	sort.SliceStable(recipes, func(i, j int) bool {
		return recipes[i].UpdatedAt.After(recipes[j].UpdatedAt)
	})
	return recipes
}

func (s *Store) GetRecipe(userID string, id string) (Recipe, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, recipe := range s.data.Recipes {
		if recipe.ID == id && recipe.UserID == userID && recipe.DeletedAt == nil {
			return recipe, nil
		}
	}
	return Recipe{}, ErrNotFound
}

func (s *Store) CreateRecipe(userID string, input RecipeMutation) (Recipe, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	input = normalizeRecipeMutation(input)
	if input.Name == "" {
		return Recipe{}, ErrInvalidInput
	}

	now := time.Now().UTC()
	recipe := Recipe{
		ID:               newID(),
		UserID:           userID,
		Name:             input.Name,
		Description:      input.Description,
		ServingCount:     input.ServingCount,
		EstimatedMinutes: input.EstimatedMinutes,
		Difficulty:       input.Difficulty,
		IsAvailable:      input.IsAvailable,
		Ingredients:      buildIngredients(input.Ingredients, nil, now),
		Steps:            buildSteps(input.Steps, nil, now),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.data.Recipes = append(s.data.Recipes, recipe)
	if err := s.saveLocked(); err != nil {
		return Recipe{}, err
	}
	return recipe, nil
}

func (s *Store) UpdateRecipe(userID string, id string, input RecipeMutation) (Recipe, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	input = normalizeRecipeMutation(input)
	if input.Name == "" {
		return Recipe{}, ErrInvalidInput
	}

	for index, recipe := range s.data.Recipes {
		if recipe.ID != id || recipe.UserID != userID || recipe.DeletedAt != nil {
			continue
		}

		now := time.Now().UTC()
		recipe.Name = input.Name
		recipe.Description = input.Description
		recipe.ServingCount = input.ServingCount
		recipe.EstimatedMinutes = input.EstimatedMinutes
		recipe.Difficulty = input.Difficulty
		recipe.IsAvailable = input.IsAvailable
		recipe.Ingredients = buildIngredients(input.Ingredients, recipe.Ingredients, now)
		recipe.Steps = buildSteps(input.Steps, recipe.Steps, now)
		recipe.UpdatedAt = now

		s.data.Recipes[index] = recipe
		if err := s.saveLocked(); err != nil {
			return Recipe{}, err
		}
		return recipe, nil
	}

	return Recipe{}, ErrNotFound
}

func (s *Store) DeleteRecipe(userID string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, recipe := range s.data.Recipes {
		if recipe.ID != id || recipe.UserID != userID || recipe.DeletedAt != nil {
			continue
		}

		now := time.Now().UTC()
		recipe.DeletedAt = &now
		recipe.UpdatedAt = now
		s.data.Recipes[index] = recipe
		return s.saveLocked()
	}
	return ErrNotFound
}

func (s *Store) IngredientSummaries(userID string) []IngredientSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := map[string]int{}
	names := map[string]string{}
	for _, recipe := range s.data.Recipes {
		if recipe.UserID != userID || recipe.DeletedAt != nil {
			continue
		}
		seenInRecipe := map[string]bool{}
		for _, ingredient := range recipe.Ingredients {
			key := strings.ToLower(strings.TrimSpace(ingredient.Name))
			if key == "" || seenInRecipe[key] {
				continue
			}
			seenInRecipe[key] = true
			names[key] = ingredient.Name
			counts[key]++
		}
	}

	items := make([]IngredientSummary, 0, len(counts))
	for key, count := range counts {
		items = append(items, IngredientSummary{Name: names[key], RecipeCount: count})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].RecipeCount == items[j].RecipeCount {
			return items[i].Name < items[j].Name
		}
		return items[i].RecipeCount > items[j].RecipeCount
	})
	return items
}

func (s *Store) saveLocked() error {
	temp := s.path + ".tmp"
	file, err := os.Create(temp)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(s.data); err != nil {
		_ = file.Close()
		_ = os.Remove(temp)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(temp)
		return err
	}
	return os.Rename(temp, s.path)
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func normalizeRecipeMutation(input RecipeMutation) RecipeMutation {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.ServingCount = clamp(input.ServingCount, 0, 99)
	input.EstimatedMinutes = clamp(input.EstimatedMinutes, 0, 24*60)
	input.Difficulty = clamp(input.Difficulty, 0, 5)
	return input
}

func buildIngredients(input []IngredientMutation, existing []Ingredient, now time.Time) []Ingredient {
	existingByID := make(map[string]Ingredient, len(existing))
	for _, ingredient := range existing {
		existingByID[ingredient.ID] = ingredient
	}

	ingredients := make([]Ingredient, 0, len(input))
	for _, item := range input {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}

		ingredient := existingByID[item.ID]
		if ingredient.ID == "" {
			ingredient.ID = newID()
			ingredient.CreatedAt = now
		}
		ingredient.Name = name
		ingredient.Amount = strings.TrimSpace(item.Amount)
		ingredient.Unit = strings.TrimSpace(item.Unit)
		ingredient.Note = strings.TrimSpace(item.Note)
		ingredient.UpdatedAt = now
		ingredients = append(ingredients, ingredient)
	}
	return ingredients
}

func buildSteps(input []StepMutation, existing []RecipeStep, now time.Time) []RecipeStep {
	existingByID := make(map[string]RecipeStep, len(existing))
	for _, step := range existing {
		existingByID[step.ID] = step
	}

	steps := make([]RecipeStep, 0, len(input))
	for index, item := range input {
		title := strings.TrimSpace(item.Title)
		description := strings.TrimSpace(item.Description)
		if title == "" && description == "" {
			continue
		}

		step := existingByID[item.ID]
		if step.ID == "" {
			step.ID = newID()
			step.CreatedAt = now
		}
		step.Title = title
		step.Description = description
		step.EstimatedMinutes = clamp(item.EstimatedMinutes, 0, 24*60)
		step.StepOrder = item.StepOrder
		if step.StepOrder <= 0 {
			step.StepOrder = index + 1
		}
		step.UpdatedAt = now
		steps = append(steps, step)
	}

	sort.SliceStable(steps, func(i, j int) bool {
		return steps[i].StepOrder < steps[j].StepOrder
	})
	for index := range steps {
		steps[index].StepOrder = index + 1
	}
	return steps
}

func clamp(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func newID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return hex.EncodeToString(bytes[0:4]) + "-" +
		hex.EncodeToString(bytes[4:6]) + "-" +
		hex.EncodeToString(bytes[6:8]) + "-" +
		hex.EncodeToString(bytes[8:10]) + "-" +
		hex.EncodeToString(bytes[10:16])
}

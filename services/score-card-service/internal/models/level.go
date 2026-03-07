package models

import "time"

// LevelPattern represents different level naming patterns
type LevelPattern string

const (
	PatternMetal        LevelPattern = "metal"         // Bronze, Silver, Gold
	PatternPerformance  LevelPattern = "performance"   // Low, Medium, High, Elite
	PatternTrafficLight LevelPattern = "traffic_light" // Red, Yellow, Orange, Green
	PatternDescriptive  LevelPattern = "descriptive"   // Basic, Good, Great
)

// Level represents a tier in a scorecard (e.g., Bronze, Silver, Gold)
type Level struct {
	ID          int64     `json:"id" db:"id"`
	ScorecardID int64     `json:"scorecard_id" db:"scorecard_id"`
	Name        string    `json:"name" db:"name"`                 // e.g., "Bronze", "Silver", "Gold"
	DisplayName string    `json:"display_name" db:"display_name"` // e.g., "🥉 Bronze"
	OrderIndex  int       `json:"order_index" db:"order_index"`   // 1=lowest, higher=better
	Color       string    `json:"color" db:"color"`               // Hex color for UI
	Icon        string    `json:"icon" db:"icon"`                 // Emoji or icon name
	Description string    `json:"description" db:"description"`
	Rules       []Rule    `json:"rules,omitempty"` // Rules for this level
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// GetLevelsByPattern returns predefined levels for a pattern
func GetLevelsByPattern(pattern LevelPattern) []Level {
	switch pattern {
	case PatternMetal:
		return []Level{
			{Name: "Starter", DisplayName: "⚪ Starter", OrderIndex: 0, Color: "#9E9E9E", Icon: "⚪", Description: "Getting started - needs improvement"},
			{Name: "Bronze", DisplayName: "🥉 Bronze", OrderIndex: 1, Color: "#CD7F32", Icon: "🥉", Description: "Basic requirements met"},
			{Name: "Silver", DisplayName: "🥈 Silver", OrderIndex: 2, Color: "#C0C0C0", Icon: "🥈", Description: "Good quality standards"},
			{Name: "Gold", DisplayName: "🥇 Gold", OrderIndex: 3, Color: "#FFD700", Icon: "🥇", Description: "Excellent quality standards"},
		}
	case PatternPerformance:
		return []Level{
			{Name: "Low", DisplayName: "Low", OrderIndex: 1, Color: "#FF6B6B", Icon: "📉"},
			{Name: "Medium", DisplayName: "Medium", OrderIndex: 2, Color: "#FFA500", Icon: "📊"},
			{Name: "High", DisplayName: "High", OrderIndex: 3, Color: "#4ECDC4", Icon: "📈"},
			{Name: "Elite", DisplayName: "🏆 Elite", OrderIndex: 4, Color: "#95E1D3", Icon: "🏆"},
		}
	case PatternTrafficLight:
		return []Level{
			{Name: "Red", DisplayName: "🔴 Red", OrderIndex: 1, Color: "#FF0000", Icon: "🔴"},
			{Name: "Yellow", DisplayName: "🟡 Yellow", OrderIndex: 2, Color: "#FFFF00", Icon: "🟡"},
			{Name: "Orange", DisplayName: "🟠 Orange", OrderIndex: 3, Color: "#FFA500", Icon: "🟠"},
			{Name: "Green", DisplayName: "🟢 Green", OrderIndex: 4, Color: "#00FF00", Icon: "🟢"},
		}
	case PatternDescriptive:
		return []Level{
			{Name: "Starter", DisplayName: "⚪ Starter", OrderIndex: 0, Color: "#9E9E9E", Icon: "⚪", Description: "Getting started"},
			{Name: "Basic", DisplayName: "⚪ Basic", OrderIndex: 1, Color: "#CCCCCC", Icon: "⚪"},
			{Name: "Good", DisplayName: "Good", OrderIndex: 2, Color: "#4ECDC4", Icon: "✅"},
			{Name: "Great", DisplayName: "Great", OrderIndex: 3, Color: "#95E1D3", Icon: "⭐"},
		}
	default:
		return []Level{}
	}
}

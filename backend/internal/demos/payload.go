package demos

import (
	"encoding/json"
)

type PublicDemoPayload struct {
	RestaurantID      string          `json:"restaurant_id"`
	RestaurantName    string          `json:"restaurant_name"`
	Cuisine           string          `json:"cuisine"`
	Hero              string          `json:"hero"`
	Hours             json.RawMessage `json:"hours,omitempty"`
	Address           string          `json:"address,omitempty"`
	Phone             string          `json:"phone,omitempty"`
	MenuSections      json.RawMessage `json:"menu_sections,omitempty"`
	ReservationCTA    string          `json:"reservation_cta,omitempty"`
	AIReceptionistCTA string          `json:"ai_receptionist_cta,omitempty"`
}

func MapPublicPayload(raw json.RawMessage) PublicDemoPayload {
	var source map[string]json.RawMessage
	if err := json.Unmarshal(raw, &source); err != nil {
		return PublicDemoPayload{}
	}

	payload := PublicDemoPayload{}
	if value, ok := source["restaurant_name"]; ok {
		_ = json.Unmarshal(value, &payload.RestaurantName)
	}
	if value, ok := source["cuisine"]; ok {
		_ = json.Unmarshal(value, &payload.Cuisine)
	}
	if value, ok := source["hero"]; ok {
		_ = json.Unmarshal(value, &payload.Hero)
	}
	if value, ok := source["hours"]; ok {
		payload.Hours = value
	}
	if value, ok := source["address"]; ok {
		_ = json.Unmarshal(value, &payload.Address)
	}
	if value, ok := source["phone"]; ok {
		_ = json.Unmarshal(value, &payload.Phone)
	}
	if value, ok := source["menu_sections"]; ok {
		payload.MenuSections = value
	}
	if value, ok := source["reservation_cta"]; ok {
		_ = json.Unmarshal(value, &payload.ReservationCTA)
	}
	if value, ok := source["ai_receptionist_cta"]; ok {
		_ = json.Unmarshal(value, &payload.AIReceptionistCTA)
	}

	return payload
}

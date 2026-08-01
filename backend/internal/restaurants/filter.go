package restaurants

import (
	"strings"

	"github.com/google/uuid"
)

func MatchesFilter(record Restaurant, filter ListFilter) bool {
	if !filter.IncludeArchived && record.Status == StatusArchived {
		return false
	}
	if filter.Restaurant != "" && !strings.Contains(strings.ToLower(record.Name), strings.ToLower(filter.Restaurant)) {
		return false
	}
	if filter.Status != "" && record.Status != filter.Status {
		return false
	}
	if filter.IsContacted != nil && record.IsContacted != *filter.IsContacted {
		return false
	}
	if filter.ShownInterest != nil && record.ShownInterest != *filter.ShownInterest {
		return false
	}
	if filter.OCRStatus != "" && record.OCRStatus != filter.OCRStatus {
		return false
	}
	return true
}

func ApplyUpdateInput(current Restaurant, input UpdateInput) Restaurant {
	updated := current
	if input.Name != nil {
		updated.Name = strings.TrimSpace(*input.Name)
	}
	if input.Email != nil {
		updated.Email = strings.TrimSpace(*input.Email)
	}
	if input.IsContacted != nil {
		updated.IsContacted = *input.IsContacted
	}
	if input.ShownInterest != nil {
		updated.ShownInterest = *input.ShownInterest
	}
	if input.Status != nil {
		updated.Status = *input.Status
	}
	return updated
}

func StatusAfterContacted(currentStatus string) string {
	switch currentStatus {
	case StatusLead, StatusDemoReady:
		return StatusEmailed
	default:
		return currentStatus
	}
}

func StatusAfterShownInterest(currentStatus string) string {
	switch currentStatus {
	case StatusLead, StatusDemoReady, StatusEmailed:
		return StatusInterested
	default:
		return currentStatus
	}
}

func filterRestaurantIDs(ids []uuid.UUID, records []Restaurant) []Restaurant {
	if len(ids) == 0 {
		return nil
	}
	allowed := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		allowed[id] = struct{}{}
	}
	filtered := make([]Restaurant, 0, len(records))
	for _, record := range records {
		if _, ok := allowed[record.ID]; ok {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

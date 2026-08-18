package scheduling

import (
	"context"
	"testing"
	"time"

	"sanitation-operations/internal/apperror"
	"sanitation-operations/internal/domain/workplan"
	"sanitation-operations/internal/pagination"
	"sanitation-operations/internal/repository"
)

type pagedShifts struct{ items []workplan.Shift }

func (p pagedShifts) GetShift(context.Context, string) (workplan.Shift, error) {
	return workplan.Shift{}, apperror.NotFound(apperror.ErrNotFound)
}
func (p pagedShifts) ListShifts(_ context.Context, _ repository.ShiftFilter, page pagination.Query) (pagination.Result[workplan.Shift], error) {
	start, end := page.Offset, page.Offset+page.Limit
	if start > len(p.items) {
		start = len(p.items)
	}
	if end > len(p.items) {
		end = len(p.items)
	}
	return pagination.Result[workplan.Shift]{Items: p.items[start:end], Total: len(p.items), Limit: page.Limit, Offset: page.Offset}, nil
}

func TestAssignmentChecksCompleteDailyRoster(t *testing.T) {
	now := time.Date(2026, 8, 18, 8, 0, 0, 0, time.UTC)
	items := make([]workplan.Shift, 101)
	for index := range items {
		items[index], _ = workplan.NewShift("existing-"+string(rune('a'+index%26))+string(rune('0'+index/26)), "route", "2026-08-18", now.Add(time.Duration(index+1)*time.Hour), now.Add(time.Duration(index+2)*time.Hour), now)
		items[index], _ = items[index].Assign("other-vehicle", now)
	}
	items[100], _ = workplan.NewShift("conflict", "route", "2026-08-18", now.Add(2*time.Hour), now.Add(3*time.Hour), now)
	items[100], _ = items[100].Assign("vehicle-1", now)
	target, _ := workplan.NewShift("target", "route", "2026-08-18", now.Add(2*time.Hour+30*time.Minute), now.Add(3*time.Hour+30*time.Minute), now)
	err := (Checker{}).CanAssign(context.Background(), pagedShifts{items: items}, target, "vehicle-1")
	if err == nil {
		t.Fatal("overlapping roster entry was not detected")
	}
}

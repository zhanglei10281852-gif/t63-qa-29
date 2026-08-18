package crew

import (
	"context"
	"errors"
	"testing"
	"time"

	"sanitation-operations/internal/audit"
	"sanitation-operations/internal/clock"
	domain "sanitation-operations/internal/domain/crew"
	"sanitation-operations/internal/identity"
	"sanitation-operations/internal/repository"
)

type certificationCancellationStore struct {
	repository.Store
	driver  domain.Driver
	started chan struct{}
}

func (s certificationCancellationStore) GetDriver(context.Context, string) (domain.Driver, error) {
	return s.driver, nil
}

func (s certificationCancellationStore) SaveDriver(context.Context, domain.Driver, int) error {
	return nil
}

func (s certificationCancellationStore) AppendAudit(ctx context.Context, _ audit.Event) error {
	close(s.started)
	<-ctx.Done()
	return ctx.Err()
}

func TestCertificationStopsWhenRequestIsCancelled(t *testing.T) {
	now := time.Date(2026, 8, 18, 8, 0, 0, 0, time.UTC)
	driver, _ := domain.New("driver-1", "DRV-901", "A", "B2", now.AddDate(1, 0, 0), now)
	started := make(chan struct{})
	service := Service{Store: certificationCancellationStore{driver: driver, started: started}, Clock: clock.Fixed{Current: now}, IDs: &identity.Sequence{}}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := service.AddCertification(ctx, CertificationInput{DriverID: driver.ID, Code: "CERT", VehicleType: "sweeper", ExpiresAt: now.AddDate(1, 0, 0)})
		done <- err
	}()
	<-started
	cancel()
	select {
	case err := <-done:
		if err == nil || !errors.Is(err, context.Canceled) {
			t.Fatalf("certification error = %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("certification did not stop after cancellation")
	}
}

package job

import (
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v2/database/model"
)

func TestShouldResetInboundMonthly_CustomDay(t *testing.T) {
	job := NewPeriodicTrafficResetJob("monthly")
	now := time.Date(2026, time.April, 15, 0, 0, 0, 0, time.UTC)

	inbound := &model.Inbound{
		TrafficReset:    "monthly",
		TrafficResetDay: 15,
	}

	if !job.shouldResetInbound(inbound, now) {
		t.Fatalf("expected inbound to reset on its configured monthly day")
	}
}

func TestShouldResetInboundMonthly_UsesLastDayForShortMonths(t *testing.T) {
	job := NewPeriodicTrafficResetJob("monthly")
	now := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)

	inbound := &model.Inbound{
		TrafficReset:    "monthly",
		TrafficResetDay: 31,
	}

	if !job.shouldResetInbound(inbound, now) {
		t.Fatalf("expected monthly reset day 31 to run on the last day of a 30-day month")
	}
}

func TestShouldResetInboundMonthly_SkipsWhenAlreadyResetThisMonth(t *testing.T) {
	job := NewPeriodicTrafficResetJob("monthly")
	now := time.Date(2026, time.April, 15, 8, 0, 0, 0, time.UTC)

	inbound := &model.Inbound{
		TrafficReset:         "monthly",
		TrafficResetDay:      15,
		LastTrafficResetTime: time.Date(2026, time.April, 15, 1, 0, 0, 0, time.UTC).UnixMilli(),
	}

	if job.shouldResetInbound(inbound, now) {
		t.Fatalf("expected inbound to skip duplicate monthly reset within the same month")
	}
}

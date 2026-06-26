package job

import (
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/logger"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service"
)

// Period represents the time period for traffic resets.
type Period string

// PeriodicTrafficResetJob resets traffic statistics for inbounds based on their configured reset period.
type PeriodicTrafficResetJob struct {
	inboundService service.InboundService
	clientService  service.ClientService
	period         Period
}

// NewPeriodicTrafficResetJob creates a new periodic traffic reset job for the specified period.
func NewPeriodicTrafficResetJob(period Period) *PeriodicTrafficResetJob {
	return &PeriodicTrafficResetJob{
		period: period,
	}
}

func (j *PeriodicTrafficResetJob) shouldResetInbound(inbound *model.Inbound, now time.Time) bool {
	if inbound == nil {
		return false
	}

	if j.period != "monthly" {
		return true
	}

	resetDay := inbound.TrafficResetDay
	if resetDay < 1 {
		resetDay = 1
	}
	if resetDay > 31 {
		resetDay = 31
	}

	lastDayOfMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
	if resetDay > lastDayOfMonth {
		resetDay = lastDayOfMonth
	}
	if now.Day() != resetDay {
		return false
	}

	if inbound.LastTrafficResetTime == 0 {
		return true
	}

	lastReset := time.UnixMilli(inbound.LastTrafficResetTime).In(now.Location())
	return lastReset.Year() != now.Year() || lastReset.Month() != now.Month()
}

// Run resets traffic statistics for all inbounds that match the configured reset period.
func (j *PeriodicTrafficResetJob) Run() {
	inbounds, err := j.inboundService.GetInboundsByTrafficReset(string(j.period))
	if err != nil {
		logger.Warning("Failed to get inbounds for traffic reset:", err)
		return
	}

	if len(inbounds) == 0 {
		return
	}
	logger.Infof("Running periodic traffic reset job for period: %s (%d matching inbounds)", j.period, len(inbounds))

	resetCount := 0
	now := time.Now()

	for _, inbound := range inbounds {
		if !j.shouldResetInbound(inbound, now) {
			continue
		}

		resetInboundErr := j.inboundService.ResetInboundTraffic(inbound.Id)
		if resetInboundErr != nil {
			logger.Warning("Failed to reset traffic for inbound", inbound.Id, ":", resetInboundErr)
		}

		resetClientErr := j.clientService.ResetAllClientTraffics(&j.inboundService, inbound.Id)
		if resetClientErr != nil {
			logger.Warning("Failed to reset traffic for all users of inbound", inbound.Id, ":", resetClientErr)
		}

		if resetInboundErr == nil && resetClientErr == nil {
			resetCount++
		}
	}

	if resetCount > 0 {
		logger.Infof("Periodic traffic reset completed: %d inbounds reset", resetCount)
	}
}

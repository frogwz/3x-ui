package service

import (
	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/xray"
)

func trafficBillingMultiplier(inbound *model.Inbound) int64 {
	if inbound != nil && inbound.EnableDoubleBilling {
		return 2
	}
	return 1
}

func trafficUsage(inbound *model.Inbound, up, down int64) int64 {
	return (up + down) * trafficBillingMultiplier(inbound)
}

func inboundTrafficUsage(inbound *model.Inbound) int64 {
	if inbound == nil {
		return 0
	}
	return trafficUsage(inbound, inbound.Up, inbound.Down)
}

func inboundAllTimeTrafficUsage(inbound *model.Inbound) int64 {
	if inbound == nil {
		return 0
	}
	base := inbound.AllTime
	if base == 0 {
		base = inbound.Up + inbound.Down
	}
	return base * trafficBillingMultiplier(inbound)
}

func clientTrafficUsage(inbound *model.Inbound, traffic *xray.ClientTraffic) int64 {
	if traffic == nil {
		return 0
	}
	return trafficUsage(inbound, traffic.Up, traffic.Down)
}

func clientRemainingTraffic(inbound *model.Inbound, traffic *xray.ClientTraffic) int64 {
	if traffic == nil {
		return 0
	}
	remaining := traffic.Total - clientTrafficUsage(inbound, traffic)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func isInboundDepleted(inbound *model.Inbound, now int64) bool {
	if inbound == nil {
		return false
	}
	if inbound.ExpiryTime > 0 && inbound.ExpiryTime <= now {
		return true
	}
	return inbound.Total > 0 && inboundTrafficUsage(inbound) >= inbound.Total
}

func isClientDepleted(inbound *model.Inbound, traffic *xray.ClientTraffic, now int64) bool {
	if traffic == nil {
		return false
	}
	if traffic.ExpiryTime > 0 && traffic.ExpiryTime <= now {
		return true
	}
	return traffic.Total > 0 && clientTrafficUsage(inbound, traffic) >= traffic.Total
}

package service

import (
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
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

// InboundBilledUsage returns the billed traffic usage for an inbound.
// When bidirectional billing is enabled on the inbound, usage = (up+down)*2.
func InboundBilledUsage(inbound *model.Inbound) int64 {
	return inboundTrafficUsage(inbound)
}

// ClientBilledUsage returns the billed traffic usage for a client,
// respecting the parent inbound's billing mode.
func ClientBilledUsage(inbound *model.Inbound, traffic *xray.ClientTraffic) int64 {
	return clientTrafficUsage(inbound, traffic)
}

// ClientBilledRemaining returns the remaining traffic for a client,
// respecting the parent inbound's billing mode.
func ClientBilledRemaining(inbound *model.Inbound, traffic *xray.ClientTraffic) int64 {
	return clientRemainingTraffic(inbound, traffic)
}

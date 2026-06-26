package service

import (
	"encoding/json"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

func TestClientTrafficUsage_DoubleBilling(t *testing.T) {
	inbound := &model.Inbound{EnableDoubleBilling: true}
	traffic := &xray.ClientTraffic{Up: 100, Down: 50}

	got := clientTrafficUsage(inbound, traffic)
	if got != 300 {
		t.Fatalf("expected double-billed client usage to be 300, got %d", got)
	}
}

func TestInboundTrafficUsage_DefaultsToSingleBilling(t *testing.T) {
	inbound := &model.Inbound{Up: 100, Down: 50}

	got := inboundTrafficUsage(inbound)
	if got != 150 {
		t.Fatalf("expected single-billed inbound usage to be 150, got %d", got)
	}
}

func TestDelDepletedClients_UsesDoubleBilling(t *testing.T) {
	setupInboundResetTestDB(t)

	settings := map[string]any{
		"clients": []map[string]any{
			{"id": "client-1", "email": "depleted@example.com", "enable": true},
			{"id": "client-2", "email": "kept@example.com", "enable": true},
		},
	}
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("marshal settings: %v", err)
	}

	inbound := &model.Inbound{
		UserId:              1,
		Remark:              "double-billing",
		Enable:              true,
		Port:                10002,
		Protocol:            model.VLESS,
		Settings:            string(settingsJSON),
		Tag:                 "inbound-double-billing",
		EnableDoubleBilling: true,
	}

	db := database.GetDB()
	if err := db.Create(inbound).Error; err != nil {
		t.Fatalf("create inbound: %v", err)
	}

	rows := []*xray.ClientTraffic{
		{
			InboundId: inbound.Id,
			Enable:    true,
			Email:     "depleted@example.com",
			Up:        50,
			Down:      40,
			Total:     150,
		},
		{
			InboundId: inbound.Id,
			Enable:    true,
			Email:     "kept@example.com",
			Up:        10,
			Down:      10,
			Total:     150,
		},
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("create client traffic %s: %v", row.Email, err)
		}
	}

	service := InboundService{}
	if err := service.DelDepletedClients(inbound.Id); err != nil {
		t.Fatalf("DelDepletedClients failed: %v", err)
	}

	var remainingInbound model.Inbound
	if err := db.First(&remainingInbound, inbound.Id).Error; err != nil {
		t.Fatalf("reload inbound: %v", err)
	}

	clients, err := service.GetClients(&remainingInbound)
	if err != nil {
		t.Fatalf("GetClients failed: %v", err)
	}
	if len(clients) != 1 || clients[0].Email != "kept@example.com" {
		t.Fatalf("expected only kept@example.com to remain, got %#v", clients)
	}

	var traffics []xray.ClientTraffic
	if err := db.Where("inbound_id = ?", inbound.Id).Order("email asc").Find(&traffics).Error; err != nil {
		t.Fatalf("reload client traffics: %v", err)
	}
	if len(traffics) != 1 || traffics[0].Email != "kept@example.com" {
		t.Fatalf("expected only kept@example.com traffic row to remain, got %#v", traffics)
	}
}

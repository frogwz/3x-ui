package service

import (
	"path/filepath"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func setupInboundResetTestDB(t *testing.T) {
	t.Helper()

	dbDir := t.TempDir()
	if err := database.InitDB(filepath.Join(dbDir, "3x-ui.db")); err != nil {
		t.Fatalf("database.InitDB failed: %v", err)
	}
	t.Cleanup(func() {
		if err := database.CloseDB(); err != nil {
			t.Logf("database.CloseDB warning: %v", err)
		}
	})
}

func TestResetAllTraffics_ResetsInboundTraffic(t *testing.T) {
	setupInboundResetTestDB(t)

	db := database.GetDB()
	inbound := &model.Inbound{
		UserId:   1,
		Up:       123,
		Down:     456,
		Remark:   "reset-all",
		Enable:   true,
		Port:     10001,
		Protocol: model.VLESS,
		Settings: `{"clients":[{"id":"client-1","email":"client@example.com","enable":true}]}`,
		Tag:      "inbound-reset-all",
	}
	if err := db.Create(inbound).Error; err != nil {
		t.Fatalf("create inbound: %v", err)
	}

	service := InboundService{}
	if err := service.ResetAllTraffics(); err != nil {
		t.Fatalf("ResetAllTraffics failed: %v", err)
	}

	var gotInbound model.Inbound
	if err := db.First(&gotInbound, inbound.Id).Error; err != nil {
		t.Fatalf("reload inbound: %v", err)
	}
	if gotInbound.Up != 0 || gotInbound.Down != 0 {
		t.Fatalf("expected inbound traffic to be reset, got up=%d down=%d", gotInbound.Up, gotInbound.Down)
	}
}

func TestResetInboundTraffic_ResetsTraffic(t *testing.T) {
	setupInboundResetTestDB(t)

	db := database.GetDB()
	inbound := &model.Inbound{
		UserId:   1,
		Up:       789,
		Down:     321,
		Remark:   "reset-one",
		Enable:   true,
		Port:     10002,
		Protocol: model.VLESS,
		Settings: `{"clients":[{"id":"client-1","email":"client@example.com","enable":true}]}`,
		Tag:      "inbound-reset-one",
	}
	if err := db.Create(inbound).Error; err != nil {
		t.Fatalf("create inbound: %v", err)
	}

	service := InboundService{}
	if err := service.ResetInboundTraffic(inbound.Id); err != nil {
		t.Fatalf("ResetInboundTraffic failed: %v", err)
	}

	var gotInbound model.Inbound
	if err := db.First(&gotInbound, inbound.Id).Error; err != nil {
		t.Fatalf("reload inbound: %v", err)
	}
	if gotInbound.Up != 0 || gotInbound.Down != 0 {
		t.Fatalf("expected inbound traffic to be reset, got up=%d down=%d", gotInbound.Up, gotInbound.Down)
	}
}

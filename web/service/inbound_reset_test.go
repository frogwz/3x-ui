package service

import (
	"path/filepath"
	"testing"

	"github.com/mhsanaei/3x-ui/v2/database"
	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/xray"
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

func TestResetAllTraffics_AlsoResetsClientTraffic(t *testing.T) {
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

	clientTraffic := &xray.ClientTraffic{
		InboundId: inbound.Id,
		Enable:    false,
		Email:     "client@example.com",
		Up:        789,
		Down:      321,
	}
	if err := db.Create(clientTraffic).Error; err != nil {
		t.Fatalf("create client traffic: %v", err)
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

	var gotClient xray.ClientTraffic
	if err := db.First(&gotClient, clientTraffic.Id).Error; err != nil {
		t.Fatalf("reload client traffic: %v", err)
	}
	if gotClient.Up != 0 || gotClient.Down != 0 {
		t.Fatalf("expected client traffic to be reset, got up=%d down=%d", gotClient.Up, gotClient.Down)
	}
	if !gotClient.Enable {
		t.Fatalf("expected client traffic to be re-enabled after reset")
	}
}

package services

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestTelegramNotificationServiceEnqueuesWeeklyPackResults(t *testing.T) {
	svc, _, adminID, packID, _, _ := newTestTelegramNotificationService(t)

	notification, err := svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{
		Type:     "weekly_pack_results",
		EntityID: packID,
		Payload:  json.RawMessage(`{"message":"results ready"}`),
	})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	if notification.Status != "queued" || notification.Type != "weekly_pack_results" || notification.EnqueuedByUserID != adminID {
		t.Fatalf("unexpected notification: %+v", notification)
	}
	if string(notification.Payload) != `{"message":"results ready"}` {
		t.Fatalf("unexpected payload: %s", string(notification.Payload))
	}
}

func TestTelegramNotificationServiceEnqueuesDefaultPayloadsForSupportedTypes(t *testing.T) {
	svc, _, adminID, _, pendingUserID, proposalID := newTestTelegramNotificationService(t)

	cases := []struct {
		name     string
		req      dto.EnqueueTelegramNotificationRequest
		fragment string
	}{
		{
			name:     "pending user",
			req:      dto.EnqueueTelegramNotificationRequest{Type: "pending_user", EntityID: pendingUserID},
			fragment: `"user_id":`,
		},
		{
			name:     "movie proposal",
			req:      dto.EnqueueTelegramNotificationRequest{Type: "movie_proposal", EntityID: proposalID},
			fragment: `"proposal_id":`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			notification, err := svc.Enqueue(adminID, tc.req)
			if err != nil {
				t.Fatalf("enqueue failed: %v", err)
			}
			if !json.Valid(notification.Payload) {
				t.Fatalf("payload is not valid json: %s", string(notification.Payload))
			}
			if !containsString(string(notification.Payload), tc.fragment) {
				t.Fatalf("expected payload %s to contain %s", string(notification.Payload), tc.fragment)
			}
		})
	}
}

func TestTelegramNotificationServiceRejectsInvalidTypeAndPayload(t *testing.T) {
	svc, _, adminID, packID, _, _ := newTestTelegramNotificationService(t)

	_, err := svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{Type: "unknown", EntityID: packID})
	if !errors.Is(err, appErrors.ErrInvalidTelegramNotificationType) {
		t.Fatalf("expected invalid type error, got %v", err)
	}

	_, err = svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{
		Type:     "weekly_pack_results",
		EntityID: packID,
		Payload:  json.RawMessage(`{bad`),
	})
	if !errors.Is(err, appErrors.ErrInvalidTelegramNotificationPayload) {
		t.Fatalf("expected invalid payload error, got %v", err)
	}

	for _, raw := range []json.RawMessage{
		json.RawMessage(`null`),
		json.RawMessage(`[]`),
		json.RawMessage(`"text"`),
		json.RawMessage(`1`),
	} {
		_, err = svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{
			Type:     "weekly_pack_results",
			EntityID: packID,
			Payload:  raw,
		})
		if !errors.Is(err, appErrors.ErrInvalidTelegramNotificationPayload) {
			t.Fatalf("expected invalid payload for %s, got %v", string(raw), err)
		}
	}
}

func TestTelegramNotificationServiceValidatesEntityExistence(t *testing.T) {
	svc, _, adminID, _, _, _ := newTestTelegramNotificationService(t)

	_, err := svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{Type: "weekly_pack_results", EntityID: 9999})
	if !errors.Is(err, appErrors.ErrWeeklyPackNotFound) {
		t.Fatalf("expected weekly pack not found, got %v", err)
	}
}

func TestTelegramNotificationServiceValidatesEntityState(t *testing.T) {
	svc, rep, adminID, _, _, _ := newTestTelegramNotificationService(t)

	votingPack := models.WeeklyPack{Name: "Voting", Status: "voting", CreatedByUserID: adminID}
	if err := rep.Create(&votingPack).Error; err != nil {
		t.Fatalf("seed voting pack: %v", err)
	}
	_, err := svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{Type: "weekly_pack_results", EntityID: votingPack.ID})
	if !errors.Is(err, appErrors.ErrInvalidTelegramNotificationEntity) {
		t.Fatalf("expected invalid weekly pack state, got %v", err)
	}

	activeUser := models.User{Username: "active-user", Password: "x", DisplayName: "Active", RoleID: 1, Status: "active"}
	if err := rep.Create(&activeUser).Error; err != nil {
		t.Fatalf("seed active user: %v", err)
	}
	_, err = svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{Type: "pending_user", EntityID: activeUser.ID})
	if !errors.Is(err, appErrors.ErrInvalidTelegramNotificationEntity) {
		t.Fatalf("expected invalid pending user state, got %v", err)
	}

	approvedProposal := models.MovieProposal{
		Title:            "Approved",
		Description:      "Long",
		SmallDescription: "Short",
		Duration:         100,
		ReleaseDate:      2020,
		Country:          "US",
		Poster:           "poster.jpg",
		RatingKp:         7,
		Source:           "manual",
		Status:           "approved",
		ProposedByUserID: activeUser.ID,
	}
	if err := rep.Create(&approvedProposal).Error; err != nil {
		t.Fatalf("seed approved proposal: %v", err)
	}
	_, err = svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{Type: "movie_proposal", EntityID: approvedProposal.ID})
	if !errors.Is(err, appErrors.ErrInvalidTelegramNotificationEntity) {
		t.Fatalf("expected invalid proposal state, got %v", err)
	}
}

func TestTelegramNotificationServiceRetriesFailedNotification(t *testing.T) {
	svc, rep, adminID, packID, _, _ := newTestTelegramNotificationService(t)

	notification, err := svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{Type: "weekly_pack_results", EntityID: packID})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}
	if err := rep.Model(&models.TelegramNotification{}).Where("id = ?", notification.ID).Updates(map[string]interface{}{
		"status":     "failed",
		"attempts":   1,
		"last_error": "network",
		"sent_at":    nil,
	}).Error; err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	retried, err := svc.Retry(notification.ID, adminID+1000)
	if err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	if retried.Status != "queued" || retried.LastError != "" || retried.Attempts != 1 {
		t.Fatalf("unexpected retried notification: %+v", retried)
	}
	if retried.EnqueuedByUserID != adminID {
		t.Fatalf("retry must not rewrite original enqueue audit, got %+v", retried)
	}
}

func TestTelegramNotificationServiceRejectsRetryForQueuedNotification(t *testing.T) {
	svc, _, adminID, packID, _, _ := newTestTelegramNotificationService(t)

	notification, err := svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{Type: "weekly_pack_results", EntityID: packID})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	_, err = svc.Retry(notification.ID, adminID)
	if !errors.Is(err, appErrors.ErrTelegramNotificationNotRetryable) {
		t.Fatalf("expected not retryable error, got %v", err)
	}
}

func TestTelegramNotificationServiceListFilters(t *testing.T) {
	svc, rep, adminID, packID, pendingUserID, proposalID := newTestTelegramNotificationService(t)

	if _, err := svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{Type: "weekly_pack_results", EntityID: packID}); err != nil {
		t.Fatalf("enqueue weekly results: %v", err)
	}
	pending, err := svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{Type: "pending_user", EntityID: pendingUserID})
	if err != nil {
		t.Fatalf("enqueue pending user: %v", err)
	}
	if _, err := svc.Enqueue(adminID, dto.EnqueueTelegramNotificationRequest{Type: "movie_proposal", EntityID: proposalID}); err != nil {
		t.Fatalf("enqueue proposal: %v", err)
	}
	if err := rep.Model(&models.TelegramNotification{}).Where("id = ?", pending.ID).Update("status", "failed").Error; err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	items, err := svc.List("failed", "pending_user")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(items) != 1 || items[0].ID != pending.ID {
		t.Fatalf("unexpected filtered list: %+v", items)
	}
}

func newTestTelegramNotificationService(t *testing.T) (TelegramNotificationService, *testRepository, uint, uint, uint, uint) {
	t.Helper()

	db := openTestSQLiteDB(t)
	rep := &testRepository{db: db}
	for _, model := range []interface{}{
		&models.Role{},
		&models.User{},
		&models.Film{},
		&models.WeeklyPack{},
		&models.MovieProposal{},
		&models.TelegramNotification{},
	} {
		if err := rep.AutoMigrate(model); err != nil {
			t.Fatalf("failed to migrate model: %v", err)
		}
	}

	if err := rep.Create(&models.Role{ID: 1, Name: "member"}).Error; err != nil {
		t.Fatalf("seed member role: %v", err)
	}
	if err := rep.Create(&models.Role{ID: 2, Name: "admin"}).Error; err != nil {
		t.Fatalf("seed admin role: %v", err)
	}

	member := models.User{Username: "member", Password: "x", DisplayName: "Member", RoleID: 1, Status: "active"}
	if err := rep.Create(&member).Error; err != nil {
		t.Fatalf("seed member: %v", err)
	}
	admin := models.User{Username: "admin", Password: "x", DisplayName: "Admin", RoleID: 2, Status: "active"}
	if err := rep.Create(&admin).Error; err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	pending := models.User{Username: "pending", Password: "x", DisplayName: "Pending", RoleID: 1, Status: "pending"}
	if err := rep.Create(&pending).Error; err != nil {
		t.Fatalf("seed pending user: %v", err)
	}

	pack := models.WeeklyPack{Name: "Week 1", Status: "closed", CreatedByUserID: admin.ID}
	if err := rep.Create(&pack).Error; err != nil {
		t.Fatalf("seed weekly pack: %v", err)
	}
	proposal := models.MovieProposal{
		Title:            "Proposal",
		Description:      "Long",
		SmallDescription: "Short",
		Duration:         100,
		ReleaseDate:      2020,
		Country:          "US",
		Poster:           "poster.jpg",
		RatingKp:         7,
		Source:           "manual",
		Status:           "pending",
		ProposedByUserID: member.ID,
	}
	if err := rep.Create(&proposal).Error; err != nil {
		t.Fatalf("seed proposal: %v", err)
	}

	return NewTelegramNotificationService(rep), rep, admin.ID, pack.ID, pending.ID, proposal.ID
}

func containsString(value, fragment string) bool {
	return len(fragment) == 0 || strings.Contains(value, fragment)
}

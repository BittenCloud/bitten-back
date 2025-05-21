package services

import (
	"bitback/internal/interfaces"
	"bitback/internal/models"
	"bitback/internal/services/dto"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type subscriptionService struct {
	subRepo  interfaces.SubscriptionRepository
	userRepo interfaces.UserRepository
}

// NewSubscriptionService creates a new instance of subscriptionService.
func NewSubscriptionService(
	subRepo interfaces.SubscriptionRepository,
	userRepo interfaces.UserRepository,
) interfaces.SubscriptionService {
	return &subscriptionService{
		subRepo:  subRepo,
		userRepo: userRepo,
	}
}

// CreateSubscription handles the creation of a new subscription.
// It validates input, calculates the end date, determines initial active status,
// and persists the subscription.
func (s *subscriptionService) CreateSubscription(ctx context.Context, input dto.CreateSubscriptionInput) (*models.Subscription, error) {
	slog.InfoContext(ctx, "CreateSubscription: attempting to create subscription", "userID", input.UserID, "plan", input.PlanName)

	// Validate user existence.
	if _, err := s.userRepo.GetByID(ctx, input.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "CreateSubscription: user not found", "userID", input.UserID)
			return nil, fmt.Errorf("user with ID %s not found", input.UserID)
		}
		slog.ErrorContext(ctx, "CreateSubscription: failed to verify user", "userID", input.UserID, "error", err)
		return nil, fmt.Errorf("failed to verify user existence: %w", err)
	}

	// Validate subscription parameters.
	if !input.DurationUnit.IsValid() || input.DurationUnit == "" {
		slog.WarnContext(ctx, "CreateSubscription: invalid duration unit", "unit", input.DurationUnit)
		return nil, fmt.Errorf("invalid or empty duration unit: '%s'", input.DurationUnit)
	}
	if input.DurationValue <= 0 {
		slog.WarnContext(ctx, "CreateSubscription: non-positive duration value", "value", input.DurationValue)
		return nil, errors.New("duration value must be positive")
	}
	if input.PlanName == "" {
		slog.WarnContext(ctx, "CreateSubscription: empty plan name")
		return nil, errors.New("plan name cannot be empty")
	}

	// Calculate the subscription's end date based on the start date and duration.
	endDate, err := calculateEndDate(input.StartDate, input.DurationUnit, input.DurationValue)
	if err != nil {
		slog.ErrorContext(ctx, "CreateSubscription: failed to calculate end date", "error", err)
		return nil, fmt.Errorf("failed to calculate end date: %w", err)
	}

	// Determine if the subscription should be initially active.
	isActive := false
	if input.PaymentStatus == "paid" && !endDate.Before(time.Now()) {
		isActive = true
	}

	// Prepare the subscription model.
	subscription := &models.Subscription{
		UserID:        input.UserID,
		PlanName:      input.PlanName,
		DurationUnit:  input.DurationUnit,
		DurationValue: input.DurationValue,
		StartDate:     input.StartDate,
		EndDate:       endDate,
		IsActive:      isActive,
		PaymentStatus: input.PaymentStatus,
		AutoRenew:     input.AutoRenew,
	}
	if input.Price != nil {
		subscription.Price = *input.Price
	}
	if input.Currency != nil {
		subscription.Currency = *input.Currency
	}

	// Save the new subscription to the repository.
	if err := s.subRepo.Create(ctx, subscription); err != nil {
		slog.ErrorContext(ctx, "CreateSubscription: failed to save subscription", "userID", input.UserID, "error", err)
		return nil, fmt.Errorf("could not create subscription: %w", err)
	}

	slog.InfoContext(ctx, "CreateSubscription: subscription created successfully", "subscriptionID", subscription.ID, "userID", input.UserID)
	return subscription, nil
}

// GetSubscriptionByID retrieves a subscription by its ID.
// The requestingUserID is used for authorization checks.
func (s *subscriptionService) GetSubscriptionByID(ctx context.Context, subscriptionID uuid.UUID, requestingUserID uuid.UUID) (*models.Subscription, error) {
	slog.InfoContext(ctx, "GetSubscriptionByID: attempting to get subscription", "subscriptionID", subscriptionID, "requestingUserID", requestingUserID)

	sub, err := s.subRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "GetSubscriptionByID: subscription not found", "subscriptionID", subscriptionID)
			return nil, fmt.Errorf("subscription with ID %s not found: %w", subscriptionID, err)
		}
		slog.ErrorContext(ctx, "GetSubscriptionByID: failed to get subscription from repo", "subscriptionID", subscriptionID, "error", err)
		return nil, fmt.Errorf("could not retrieve subscription: %w", err)
	}

	if sub.UserID != requestingUserID {
		// TODO: Implement role-based access control for administrators.
		slog.WarnContext(ctx, "GetSubscriptionByID: user not authorized to view this subscription", "subscriptionID", subscriptionID, "subscriptionUserID", sub.UserID, "requestingUserID", requestingUserID)
		return nil, fmt.Errorf("user not authorized to view subscription %s", subscriptionID)
	}

	slog.InfoContext(ctx, "GetSubscriptionByID: subscription retrieved successfully", "subscriptionID", sub.ID)
	return sub, nil
}

// ListUserSubscriptions retrieves a paginated list of subscriptions for a specific user.
func (s *subscriptionService) ListUserSubscriptions(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Subscription, int64, error) {
	slog.InfoContext(ctx, "ListUserSubscriptions: listing subscriptions for user", "userID", userID, "page", page, "pageSize", pageSize)

	// Apply default pagination parameters if necessary.
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	offset := (page - 1) * pageSize

	subs, totalCount, err := s.subRepo.ListByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		slog.ErrorContext(ctx, "ListUserSubscriptions: failed to list subscriptions from repo", "userID", userID, "error", err)
		return nil, 0, fmt.Errorf("could not retrieve user subscriptions: %w", err)
	}
	slog.InfoContext(ctx, "ListUserSubscriptions: subscriptions listed successfully", "userID", userID, "count", len(subs), "totalCount", totalCount)
	return subs, totalCount, nil
}

// CancelSubscription handles the cancellation of a subscription.
// This typically involves disabling auto-renewal and potentially deactivating the subscription.
// The requestingUserID is used for authorization.
func (s *subscriptionService) CancelSubscription(ctx context.Context, subscriptionID uuid.UUID, requestingUserID uuid.UUID) (*models.Subscription, error) {
	slog.InfoContext(ctx, "CancelSubscription: attempting to cancel subscription", "subscriptionID", subscriptionID, "requestingUserID", requestingUserID)

	sub, err := s.subRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subscription %s not found: %w", subscriptionID, err)
		}
		return nil, fmt.Errorf("could not retrieve subscription to cancel: %w", err)
	}

	// Authorization check.
	if sub.UserID != requestingUserID {
		// TODO: Implement role-based access control for administrators.
		return nil, fmt.Errorf("user not authorized to cancel subscription %s", subscriptionID)
	}

	if !sub.IsActive && sub.EndDate.Before(time.Now()) {
		slog.InfoContext(ctx, "CancelSubscription: subscription already inactive and ended", "subscriptionID", subscriptionID)
	}

	sub.AutoRenew = false

	if err := s.subRepo.Update(ctx, sub); err != nil {
		slog.ErrorContext(ctx, "CancelSubscription: failed to update subscription for cancellation", "subscriptionID", subscriptionID, "error", err)
		return nil, fmt.Errorf("could not save subscription cancellation: %w", err)
	}

	slog.InfoContext(ctx, "CancelSubscription: subscription cancelled (auto-renew disabled)", "subscriptionID", sub.ID)
	return sub, nil
}

// UpdatePaymentStatus updates the payment status of a subscription.
// This might be invoked by a payment gateway or an administrator.
func (s *subscriptionService) UpdatePaymentStatus(ctx context.Context, subscriptionID uuid.UUID, paymentStatus string) (*models.Subscription, error) {
	slog.InfoContext(ctx, "UpdatePaymentStatus: attempting to update payment status", "subscriptionID", subscriptionID, "newStatus", paymentStatus)
	sub, err := s.subRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve subscription to update payment status: %w", err)
	}

	sub.PaymentStatus = paymentStatus
	if paymentStatus == "paid" && !sub.StartDate.After(time.Now()) && sub.EndDate.After(time.Now()) {
		sub.IsActive = true
	} else if paymentStatus == "failed" || paymentStatus == "refunded" {
		sub.IsActive = false
	}

	if err := s.subRepo.Update(ctx, sub); err != nil {
		slog.ErrorContext(ctx, "UpdatePaymentStatus: failed to save subscription payment status", "subscriptionID", subscriptionID, "error", err)
		return nil, fmt.Errorf("could not save subscription payment status: %w", err)
	}
	slog.InfoContext(ctx, "UpdatePaymentStatus: payment status updated", "subscriptionID", sub.ID, "newStatus", sub.PaymentStatus)
	return sub, nil
}

// SetAutoRenew sets the auto-renewal flag for a subscription.
// The requestingUserID is used for authorization.
func (s *subscriptionService) SetAutoRenew(ctx context.Context, subscriptionID uuid.UUID, requestingUserID uuid.UUID, autoRenew bool) (*models.Subscription, error) {
	slog.InfoContext(ctx, "SetAutoRenew: setting auto-renew status", "subscriptionID", subscriptionID, "autoRenew", autoRenew, "requestingUserID", requestingUserID)
	sub, err := s.subRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve subscription: %w", err)
	}

	// Authorization check.
	if sub.UserID != requestingUserID {
		// TODO: Implement role-based access control for administrators.
		return nil, fmt.Errorf("user not authorized to set auto-renew for subscription %s", subscriptionID)
	}

	if sub.AutoRenew == autoRenew {
		slog.InfoContext(ctx, "SetAutoRenew: auto-renew status already set to desired value", "subscriptionID", subscriptionID, "autoRenew", autoRenew)
		return sub, nil
	}

	sub.AutoRenew = autoRenew
	if err := s.subRepo.Update(ctx, sub); err != nil {
		slog.ErrorContext(ctx, "SetAutoRenew: failed to update auto-renew status", "subscriptionID", subscriptionID, "error", err)
		return nil, fmt.Errorf("could not save auto-renew status: %w", err)
	}
	slog.InfoContext(ctx, "SetAutoRenew: auto-renew status updated", "subscriptionID", sub.ID, "autoRenew", sub.AutoRenew)
	return sub, nil
}

// GetUsersWithExpiringSubscriptions retrieves users and their subscriptions that are nearing expiration.
// The report is paginated based on the subscriptions, not directly on users.
func (s *subscriptionService) GetUsersWithExpiringSubscriptions(ctx context.Context, daysInAdvance int, page, pageSize int) ([]dto.UserWithExpiringSubscriptions, int64, error) {
	slog.InfoContext(ctx, "GetUsersWithExpiringSubscriptions: fetching report", "daysInAdvance", daysInAdvance, "page", page, "pageSize", pageSize)

	if daysInAdvance < 0 {
		daysInAdvance = 0 // Consider subscriptions expiring from now onwards.
	}
	// Apply default pagination parameters.
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	now := time.Now()
	thresholdDateFrom := now // Subscriptions expiring from the current moment.
	thresholdDateTo := now.AddDate(0, 0, daysInAdvance)
	offset := (page - 1) * pageSize // Pagination applies to the list of expiring subscriptions.

	// Retrieve all expiring subscriptions within the date range, with pagination.
	expiringSubs, totalExpiringSubsCount, err := s.subRepo.ListExpiringSoon(ctx, thresholdDateFrom, thresholdDateTo, offset, pageSize)
	if err != nil {
		slog.ErrorContext(ctx, "GetUsersWithExpiringSubscriptions: failed to list expiring subscriptions", "error", err)
		return nil, 0, fmt.Errorf("could not list expiring subscriptions: %w", err)
	}

	if len(expiringSubs) == 0 {
		return []dto.UserWithExpiringSubscriptions{}, 0, nil
	}

	// Collect unique UserIDs from the retrieved subscriptions.
	userIDsMap := make(map[uuid.UUID]bool)
	for _, sub := range expiringSubs {
		userIDsMap[sub.UserID] = true
	}
	uniqueUserIDs := make([]uuid.UUID, 0, len(userIDsMap))
	for uid := range userIDsMap {
		uniqueUserIDs = append(uniqueUserIDs, uid)
	}

	// Fetch all associated users in a single query.
	users, err := s.userRepo.GetByIDs(ctx, uniqueUserIDs)
	if err != nil {
		slog.ErrorContext(ctx, "GetUsersWithExpiringSubscriptions: failed to get users by IDs", "error", err)
		return nil, 0, fmt.Errorf("could not fetch users for expiring subscriptions: %w", err)
	}

	// Group subscriptions by user for the report.
	usersMap := make(map[uuid.UUID]models.User)
	for _, u := range users {
		usersMap[u.ID] = u
	}

	reportDataMap := make(map[uuid.UUID]*dto.UserWithExpiringSubscriptions)
	for _, sub := range expiringSubs {
		user, ok := usersMap[sub.UserID]
		if !ok {
			// This case might occur if a user was deleted after their subscription was fetched.
			slog.WarnContext(ctx, "GetUsersWithExpiringSubscriptions: user not found for subscription, skipping", "userID", sub.UserID, "subscriptionID", sub.ID)
			continue
		}

		if _, exists := reportDataMap[user.ID]; !exists {
			reportDataMap[user.ID] = &dto.UserWithExpiringSubscriptions{
				User:                  user,
				ExpiringSubscriptions: []dto.ExpiringSubscriptionInfo{},
			}
		}
		reportDataMap[user.ID].ExpiringSubscriptions = append(reportDataMap[user.ID].ExpiringSubscriptions, dto.ExpiringSubscriptionInfo{
			ID:            sub.ID,
			PlanName:      sub.PlanName,
			EndDate:       sub.EndDate,
			DurationUnit:  sub.DurationUnit,
			DurationValue: sub.DurationValue,
			AutoRenew:     sub.AutoRenew,
		})
	}

	// Convert the map to a slice for the response.
	// The totalExpiringSubsCount refers to the total number of expiring *subscriptions*, not unique users.
	finalReportData := make([]dto.UserWithExpiringSubscriptions, 0, len(reportDataMap))
	for _, data := range reportDataMap {
		finalReportData = append(finalReportData, *data)
	}

	slog.InfoContext(ctx, "GetUsersWithExpiringSubscriptions: report generated", "userCountInPage", len(finalReportData), "totalExpiringSubscriptionsAcrossAllPages", totalExpiringSubsCount)
	return finalReportData, totalExpiringSubsCount, nil
}

// ListActiveSubscriptionsByPlan retrieves a paginated list of active subscriptions for a specific plan name.
func (s *subscriptionService) ListActiveSubscriptionsByPlan(ctx context.Context, planName string, page, pageSize int) ([]models.Subscription, int64, error) {
	slog.InfoContext(ctx, "ListActiveSubscriptionsByPlan: listing active subscriptions", "planName", planName, "page", page, "pageSize", pageSize)

	if strings.TrimSpace(planName) == "" {
		return nil, 0, errors.New("plan name cannot be empty")
	}

	// Apply default pagination parameters.
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	offset := (page - 1) * pageSize

	subs, totalCount, err := s.subRepo.ListActiveByPlanName(ctx, planName, offset, pageSize)
	if err != nil {
		slog.ErrorContext(ctx, "ListActiveSubscriptionsByPlan: failed to list subscriptions from repo", "planName", planName, "error", err)
		return nil, 0, fmt.Errorf("could not retrieve active subscriptions for plan '%s': %w", planName, err)
	}

	slog.InfoContext(ctx, "ListActiveSubscriptionsByPlan: subscriptions listed successfully", "planName", planName, "count", len(subs), "totalCount", totalCount)
	return subs, totalCount, nil
}

// CheckUserActiveSubscription checks if a user has any active subscription.
func (s *subscriptionService) CheckUserActiveSubscription(ctx context.Context, userID uuid.UUID) (bool, error) {
	slog.InfoContext(ctx, "CheckUserActiveSubscription: checking active subscription", "userID", userID)
	hasActiveSub, err := s.subRepo.CheckUserActiveSubscription(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "CheckUserActiveSubscription: failed to check subscription status from repo", "userID", userID, "error", err)
		return false, fmt.Errorf("could not check user's active subscription: %w", err)
	}
	slog.InfoContext(ctx, "CheckUserActiveSubscription: status checked", "userID", userID, "hasActiveSubscription", hasActiveSub)
	return hasActiveSub, nil
}

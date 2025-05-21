package services

import (
	"bitback/internal/interfaces"
	"bitback/internal/models"
	"bitback/internal/services/dto" // Assuming you created keyServiceDTO.go
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type keyService struct {
	userRepo         interfaces.UserRepository
	hostRepo         interfaces.HostRepository
	subscriptionRepo interfaces.SubscriptionRepository
}

// NewKeyService creates a new instance of KeyService.
func NewKeyService(ur interfaces.UserRepository, hr interfaces.HostRepository, sr interfaces.SubscriptionRepository) interfaces.KeyService {
	return &keyService{
		userRepo:         ur,
		hostRepo:         hr,
		subscriptionRepo: sr,
	}
}

// GenerateVlessKeyForUser generates a VLESS key string for a given user.
// It selects an active host based on subscription status and constructs the VLESS URL.
func (s *keyService) GenerateVlessKeyForUser(ctx context.Context, userID uuid.UUID, remarks string, country *string) (*dto.GenerateUserKeyResult, error) {
	slog.InfoContext(ctx, "GenerateVlessKeyForUser: attempting to generate key", "userID", userID, "country", country)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "GenerateVlessKeyForUser: user not found", "userID", userID)
			return nil, fmt.Errorf("user with ID %s not found", userID)
		}
		slog.ErrorContext(ctx, "GenerateVlessKeyForUser: failed to get user", "userID", userID, "error", err)
		return nil, fmt.Errorf("could not retrieve user: %w", err)
	}

	hasActiveSubscription, err := s.subscriptionRepo.CheckUserActiveSubscription(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "GenerateVlessKeyForUser: failed to check user subscription status", "userID", userID, "error", err)
		hasActiveSubscription = false // Default to no subscription if check fails
	}

	var hostTier bool // true for free, false for paid
	if hasActiveSubscription {
		slog.InfoContext(ctx, "GenerateVlessKeyForUser: user has active subscription, seeking paid host", "userID", userID)
		hostTier = false // User has subscription, look for a paid host
	} else {
		slog.InfoContext(ctx, "GenerateVlessKeyForUser: user has no active subscription, seeking free host", "userID", userID)
		hostTier = true // User has no subscription, look for a free host
	}

	host, err := s.hostRepo.GetRandomActiveHost(ctx, country, &hostTier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "GenerateVlessKeyForUser: no active hosts available for the tier/country", "tier_is_free", hostTier, "country", country)
			// Try fallback: if a specific country was requested and no host found, try without country filter for the same tier
			if country != nil && *country != "" {
				slog.InfoContext(ctx, "GenerateVlessKeyForUser: fallback - trying without country filter for tier", "tier_is_free", hostTier)
				host, err = s.hostRepo.GetRandomActiveHost(ctx, nil, &hostTier)
			}
		}
		// If still not found or other error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.WarnContext(ctx, "GenerateVlessKeyForUser: no active hosts available even after fallback", "tier_is_free", hostTier)
				return nil, errors.New("no active hosts available to generate key for the specified criteria")
			}
			slog.ErrorContext(ctx, "GenerateVlessKeyForUser: failed to get active host", "error", err)
			return nil, fmt.Errorf("could not retrieve an active host: %w", err)
		}
	}
	slog.DebugContext(ctx, "GenerateVlessKeyForUser: selected host", "hostID", host.ID, "hostAddress", host.Address, "isFreeTier", host.IsFreeTier)

	vlessUserID := user.ID.String()
	vlessURL, err := s.constructVlessURL(vlessUserID, host, remarks)
	if err != nil {
		slog.ErrorContext(ctx, "GenerateVlessKeyForUser: failed to construct VLESS URL", "userID", userID, "hostID", host.ID, "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "GenerateVlessKeyForUser: VLESS key generated successfully", "userID", userID, "hostID", host.ID, "hasActiveSubscription", hasActiveSubscription)
	return &dto.GenerateUserKeyResult{
		VlessKey:              vlessURL,
		HasActiveSubscription: hasActiveSubscription,
	}, nil
}

// GenerateFreeVlessKey generates a VLESS key for a free-tier user.
func (s *keyService) GenerateFreeVlessKey(ctx context.Context, remarks string, country *string) (string, error) {
	slog.InfoContext(ctx, "GenerateFreeVlessKey: attempting to generate free key", "country", country)

	isFreeHost := true
	host, err := s.hostRepo.GetRandomActiveHost(ctx, country, &isFreeHost)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "GenerateFreeVlessKey: no active free hosts available for the country", "country", country)
			// Try fallback: if a specific country was requested and no host found, try without country filter for free tier
			if country != nil && *country != "" {
				slog.InfoContext(ctx, "GenerateFreeVlessKey: fallback - trying without country filter for free tier")
				host, err = s.hostRepo.GetRandomActiveHost(ctx, nil, &isFreeHost)
			}
		}
		// If still not found or other error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.WarnContext(ctx, "GenerateFreeVlessKey: no active free hosts available even after fallback")
				return "", errors.New("no active free hosts available to generate key")
			}
			slog.ErrorContext(ctx, "GenerateFreeVlessKey: failed to get active free host", "error", err)
			return "", fmt.Errorf("could not retrieve an active free host: %w", err)
		}
	}
	slog.DebugContext(ctx, "GenerateFreeVlessKey: selected host", "hostID", host.ID, "hostAddress", host.Address)

	vlessURL, err := s.constructVlessURL(FreeTierUserUUID.String(), host, remarks)
	if err != nil {
		slog.ErrorContext(ctx, "GenerateFreeVlessKey: failed to construct VLESS URL", "hostID", host.ID, "error", err)
		return "", err
	}

	slog.InfoContext(ctx, "GenerateFreeVlessKey: VLESS key generated successfully", "hostID", host.ID)
	return vlessURL, nil
}

// constructVlessURL is a helper function to build the VLESS URL string.
func (s *keyService) constructVlessURL(vlessUserID string, host *models.Host, remarks string) (string, error) {
	queryParams := url.Values{}

	if host.SecurityType != "" && host.SecurityType != "none" {
		queryParams.Set("security", host.SecurityType)
	}
	if host.SNI != "" {
		queryParams.Set("sni", host.SNI)
	}
	if host.Fingerprint != "" {
		queryParams.Set("fp", host.Fingerprint)
	}

	if strings.ToLower(host.SecurityType) == "reality" {
		if host.PublicKey == "" {
			return "", fmt.Errorf("selected host (ID: %d) is configured for Reality but missing public key (pbk)", host.ID)
		}
		queryParams.Set("pbk", host.PublicKey)
		if host.RSID != "" {
			queryParams.Set("sid", host.RSID)
		}
	}

	if host.Flow != "" {
		queryParams.Set("flow", host.Flow)
	}

	if host.Network != "" {
		queryParams.Set("type", host.Network)
	} else {
		queryParams.Set("type", "tcp") // Default to tcp if not specified
	}

	queryString := queryParams.Encode()

	var vlessURL string
	if queryString != "" {
		vlessURL = fmt.Sprintf("vless://%s@%s:%s?%s", vlessUserID, host.Address, host.Port, queryString)
	} else {
		vlessURL = fmt.Sprintf("vless://%s@%s:%s", vlessUserID, host.Address, host.Port)
	}

	if remarks != "" {
		vlessURL = fmt.Sprintf("%s#%s", vlessURL, url.PathEscape(remarks))
	}
	return vlessURL, nil
}

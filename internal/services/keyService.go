package services

import (
	"bitback/internal/interfaces"
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
	userRepo interfaces.UserRepository
	hostRepo interfaces.HostRepository
}

// NewKeyService creates a new instance of KeyService.
func NewKeyService(ur interfaces.UserRepository, hr interfaces.HostRepository) interfaces.KeyService {
	return &keyService{
		userRepo: ur,
		hostRepo: hr,
	}
}

// GenerateVlessKeyForUser generates a VLESS key string for a given user.
// It selects an active host and constructs the VLESS URL with appropriate parameters.
func (s *keyService) GenerateVlessKeyForUser(ctx context.Context, userID uuid.UUID, remarks string) (string, error) {
	slog.InfoContext(ctx, "GenerateVlessKeyForUser: attempting to generate key", "userID", userID)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "GenerateVlessKeyForUser: user not found", "userID", userID)
			return "", fmt.Errorf("user with ID %s not found", userID)
		}
		slog.ErrorContext(ctx, "GenerateVlessKeyForUser: failed to get user", "userID", userID, "error", err)
		return "", fmt.Errorf("could not retrieve user: %w", err)
	}

	host, err := s.hostRepo.GetRandomActiveHost(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "GenerateVlessKeyForUser: no active hosts available")
			return "", errors.New("no active hosts available to generate key")
		}
		slog.ErrorContext(ctx, "GenerateVlessKeyForUser: failed to get active host", "error", err)
		return "", fmt.Errorf("could not retrieve an active host: %w", err)
	}
	slog.DebugContext(ctx, "GenerateVlessKeyForUser: selected host", "hostID", host.ID, "hostAddress", host.Address)

	vlessUserID := user.ID.String()
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

	// For "reality" security type, public key (pbk) is mandatory,
	// and RSID (sid) is used if available.
	if strings.ToLower(host.SecurityType) == "reality" {
		if host.PublicKey == "" {
			slog.WarnContext(ctx, "GenerateVlessKeyForUser: host selected for reality has no public key (pbk)", "hostID", host.ID)
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

	// Set the network type, defaulting to "tcp" if not specified.
	if host.Network != "" {
		queryParams.Set("type", host.Network)
	} else {
		queryParams.Set("type", "tcp")
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

	slog.InfoContext(ctx, "GenerateVlessKeyForUser: VLESS key generated successfully", "userID", userID, "hostID", host.ID)
	return vlessURL, nil
}

func (s *keyService) GenerateFreeVlessKey(ctx context.Context, remarks string) (string, error) {
	slog.InfoContext(ctx, "GenerateFreeVlessKey: attempting to generate key")

	host, err := s.hostRepo.GetRandomActiveHost(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "GenerateVlessKeyForUser: no active hosts available")
			return "", errors.New("no active hosts available to generate key")
		}
		slog.ErrorContext(ctx, "GenerateFreeVlessKey: failed to get active host", "error", err)
		return "", fmt.Errorf("could not retrieve an active host: %w", err)
	}
	slog.DebugContext(ctx, "GenerateFreeVlessKey: selected host", "hostID", host.ID, "hostAddress", host.Address)

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

	// For "reality" security type, public key (pbk) is mandatory,
	// and RSID (sid) is used if available.
	if strings.ToLower(host.SecurityType) == "reality" {
		if host.PublicKey == "" {
			slog.WarnContext(ctx, "GenerateFreeVlessKey: host selected for reality has no public key (pbk)", "hostID", host.ID)
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

	// Set the network type, defaulting to "tcp" if not specified.
	if host.Network != "" {
		queryParams.Set("type", host.Network)
	} else {
		queryParams.Set("type", "tcp")
	}

	queryString := queryParams.Encode()

	freeUserID := "0196ecf3-2557-7aeb-8593-6099ee8cb84d"

	var vlessURL string
	if queryString != "" {
		vlessURL = fmt.Sprintf("vless://%s@%s:%s?%s", freeUserID, host.Address, host.Port, queryString)
	} else {
		vlessURL = fmt.Sprintf("vless://%s@%s:%s", freeUserID, host.Address, host.Port)
	}

	if remarks != "" {
		vlessURL = fmt.Sprintf("%s#%s", vlessURL, url.PathEscape(remarks))
	}

	slog.InfoContext(ctx, "GenerateFreeVlessKey: VLESS key generated successfully", "hostID", host.ID)
	return vlessURL, nil
}

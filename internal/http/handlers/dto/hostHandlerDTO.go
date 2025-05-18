package dto

import (
	"bitback/internal/models/customTypes"
	"time"
)

// CreateHostRequest defines the request body for creating a new host.
type CreateHostRequest struct {
	HostName     string `json:"host_name,omitempty"`                                     // Optional: A descriptive name for the host.
	Country      string `json:"country,omitempty" validate:"omitempty,iso3166_1_alpha2"` // Optional: ISO 3166-1 alpha-2 country code.
	City         string `json:"city,omitempty"`                                          // Optional: City where the host is located.
	Address      string `json:"address" validate:"required"`                             // Mandatory: IP address or domain name of the host.
	Port         string `json:"port" validate:"required,numeric"`                        // Mandatory: Port number for the host service.
	Protocol     string `json:"protocol" validate:"required"`                            // Mandatory: Protocol (e.g., http, https, tcp).
	Network      string `json:"network,omitempty" validate:"omitempty"`                  // Optional: Network type (e.g., tcp, ws, grpc); can have a default in the database or service.
	PublicKey    string `json:"public_key,omitempty" validate:"omitempty"`               // Optional: Public key, used for certain security types like Reality.
	Flow         string `json:"flow,omitempty"`                                          // Optional: Flow control mechanism.
	RSID         string `json:"rsid,omitempty"`                                          // Optional: Reality Short ID.
	SecurityType string `json:"security_type,omitempty"`                                 // Optional: Security type (e.g., tls, none, reality).
	SNI          string `json:"sni,omitempty"`                                           // Optional: Server Name Indication for TLS.
	Fingerprint  string `json:"fingerprint,omitempty"`                                   // Optional: TLS fingerprint.
	IsPrivate    bool   `json:"is_private,omitempty"`                                    // Optional: Specifies if the host is private; defaults to false if omitted.
	Region       string `json:"region,omitempty"`                                        // Optional: Geographical or logical region of the host.
	Provider     string `json:"provider,omitempty"`                                      // Optional: Provider or owner of the host infrastructure.
}

// UpdateHostRequest defines the request body for updating an existing host.
// Pointer fields are used to differentiate between zero values and fields not provided for update.
type UpdateHostRequest struct {
	HostName     *string `json:"host_name,omitempty"`
	Country      *string `json:"country,omitempty" validate:"omitempty,iso3166_1_alpha2"`
	City         *string `json:"city,omitempty"`
	Address      *string `json:"address,omitempty"`                      // Typically not changed or requires special handling.
	Port         *string `json:"port,omitempty"`                         // Typically not changed or requires special handling.
	Protocol     *string `json:"protocol,omitempty"`                     // Typically not changed or requires special handling.
	Network      *string `json:"network,omitempty" validate:"omitempty"` // Network type.
	PublicKey    *string `json:"public_key,omitempty" validate:"omitempty"`
	Flow         *string `json:"flow,omitempty"`
	RSID         *string `json:"rsid,omitempty"`
	SecurityType *string `json:"security_type,omitempty"`
	SNI          *string `json:"sni,omitempty"`
	Fingerprint  *string `json:"fingerprint,omitempty"`
	IsPrivate    *bool   `json:"is_private,omitempty"`
	Region       *string `json:"region,omitempty"`
	Provider     *string `json:"provider,omitempty"`
}

// UpdateHostStatusRequest defines the request body for updating a host's online status.
type UpdateHostStatusRequest struct {
	IsOnline bool                   `json:"is_online"`                  // The new online status of the host.
	Status   customTypes.HostStatus `json:"status" validate:"required"` // The new detailed status of the host; must be a valid HostStatus.
}

// HostResponse defines the standard API response for a single host.
type HostResponse struct {
	ID            uint                   `json:"id"`
	HostName      string                 `json:"host_name,omitempty"`
	Country       string                 `json:"country,omitempty"`
	City          string                 `json:"city,omitempty"`
	Address       string                 `json:"address"`
	Port          string                 `json:"port"`
	Protocol      string                 `json:"protocol"`
	Network       string                 `json:"network,omitempty"` // Network type.
	PublicKey     string                 `json:"public_key,omitempty"`
	Flow          string                 `json:"flow,omitempty"`
	RSID          string                 `json:"rsid,omitempty"`
	SecurityType  string                 `json:"security_type,omitempty"`
	SNI           string                 `json:"sni,omitempty"`
	Fingerprint   string                 `json:"fingerprint,omitempty"`
	IsPrivate     bool                   `json:"is_private"`
	IsOnline      bool                   `json:"is_online"`
	Status        customTypes.HostStatus `json:"status"` // HostStatus will be serialized to its string representation.
	LastCheckedAt *time.Time             `json:"last_checked_at,omitempty"`
	Region        string                 `json:"region,omitempty"`
	Provider      string                 `json:"provider,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// PaginatedHostsResponse defines the structure for a paginated list of hosts.
type PaginatedHostsResponse struct {
	Hosts       []HostResponse `json:"hosts"`        // Slice of host responses for the current page.
	TotalItems  int64          `json:"total_items"`  // Total number of hosts matching the query.
	TotalPages  int            `json:"total_pages"`  // Total number of pages available.
	CurrentPage int            `json:"current_page"` // The current page number.
	PageSize    int            `json:"page_size"`    // The number of items per page.
}

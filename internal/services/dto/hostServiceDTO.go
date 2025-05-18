package dto

import (
	"bitback/internal/models/customTypes"
)

// CreateHostInput defines the data required to create a new host at the service layer.
type CreateHostInput struct {
	HostName     string // Optional: A descriptive name for the host.
	Country      string // Optional: The country where the host is located.
	City         string // Optional: The city where the host is located.
	Address      string // Mandatory: The IP address or domain name of the host.
	Port         string // Mandatory: The port number for the host service.
	Protocol     string // Mandatory: The protocol used by the host service (e.g., http, https, tcp).
	Network      string // Optional: The network type (e.g., tcp, ws, grpc); defaults to "tcp" if not specified or handled by service logic.
	PublicKey    string // Optional: The public key, often used for specific security protocols (e.g., Reality).
	Flow         string // Optional: Flow control mechanism or specific protocol feature.
	RSID         string // Optional: Reality Short ID.
	SecurityType string // Optional: The security type (e.g., tls, none, reality).
	SNI          string // Optional: Server Name Indication, used in TLS.
	Fingerprint  string // Optional: TLS fingerprint or similar identifier.
	IsPrivate    bool   // Specifies if the host is private; defaults to false.
	Region       string // Optional: The geographical or logical region of the host.
	Provider     string // Optional: The provider or owner of the host infrastructure.
}

// UpdateHostInput defines the data for updating an existing host at the service layer.
// Fields are pointers to distinguish between zero values and fields not provided for update.
type UpdateHostInput struct {
	HostName     *string // A descriptive name for the host.
	Country      *string // The country where the host is located.
	City         *string // The city where the host is located.
	Address      *string // The IP address or domain name; changing this might require special handling or re-verification.
	Port         *string // The port number; changing this might require special handling or re-verification.
	Protocol     *string // The protocol; changing this might require special handling or re-verification.
	Network      *string // The network type (e.g., tcp, ws, grpc).
	PublicKey    *string // The public key.
	Flow         *string // Flow control mechanism.
	RSID         *string // Reality Short ID.
	SecurityType *string // The security type (e.g., tls, none).
	SNI          *string // Server Name Indication.
	Fingerprint  *string // TLS fingerprint.
	IsPrivate    *bool   // Specifies if the host is private.
	Region       *string // The geographical or logical region of the host.
	Provider     *string // The provider or owner of the host infrastructure.
	// Note: IsOnline, Status, and LastCheckedAt are typically updated via separate mechanisms (e.g., monitoring).
}

// ListHostsServiceParams defines parameters for listing hosts at the service layer.
// These are subsequently mapped to repository-level parameters.
type ListHostsServiceParams struct {
	Page      int
	PageSize  int
	Country   *string
	City      *string
	Protocol  *string
	Network   *string // Filter by network type.
	IsOnline  *bool
	IsPrivate *bool
	Status    *customTypes.HostStatus // Filter by host status, using a pointer to allow omitting this filter.
	HostName  *string                 // Filter by partial host name match.
	Address   *string                 // Filter by partial address match.
	SortBy    string                  // Field to sort by (e.g., "created_at", "host_name").
	SortOrder string                  // Sort order ("asc" or "desc").
}

// UpdateHostStatusInput defines the data for specifically updating a host's online status.
type UpdateHostStatusInput struct {
	IsOnline bool                   // The new online status.
	Status   customTypes.HostStatus // The new detailed status; not a pointer as it should be explicitly set.
}

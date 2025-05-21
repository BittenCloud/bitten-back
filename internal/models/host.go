package models

import (
	"bitback/internal/models/customTypes"
	"gorm.io/gorm"
	"time"
)

// Host defines the database model for a host or server.
type Host struct {
	ID            uint                   `gorm:"primaryKey" json:"id"`
	HostName      string                 `json:"host_name,omitempty" gorm:"index"`                               // Optional: A descriptive name for the host.
	Country       string                 `json:"country,omitempty" gorm:"index"`                                 // Optional: The country where the host is located.
	City          string                 `json:"city,omitempty" gorm:"index"`                                    // Optional: The city where the host is located.
	Region        string                 `json:"region,omitempty" gorm:"index"`                                  // Optional: The geographical or logical region of the host.
	Provider      string                 `json:"provider,omitempty"`                                             // Optional: The provider or owner of the host infrastructure.
	Address       string                 `json:"address" gorm:"not null;"`                                       // Mandatory: The IP address or domain name of the host.
	Port          string                 `json:"port" gorm:"not null;"`                                          // Mandatory: The port number for the host service.
	Protocol      string                 `json:"protocol" gorm:"type:varchar(10);not null;"`                     // Mandatory: The protocol (e.g., http, https, tcp).
	Network       string                 `json:"network,omitempty" gorm:"type:varchar(10);default:'tcp';index;"` // Network type (e.g., tcp, ws, grpc, kcp). Defaults to 'tcp'.
	PublicKey     string                 `json:"public_key,omitempty" gorm:"type:text"`                          // Public key, often used for specific security protocols (e.g., Reality).
	Flow          string                 `json:"flow,omitempty"`                                                 // Flow control mechanism or specific protocol feature.
	RSID          string                 `json:"rsid,omitempty" gorm:"column:rsid"`                              // Reality Short ID.
	SecurityType  string                 `json:"security_type,omitempty"`                                        // Security type (e.g., tls, none, reality).
	SNI           string                 `json:"sni,omitempty" gorm:"column:sni"`                                // Server Name Indication, used in TLS.
	Fingerprint   string                 `json:"fingerprint,omitempty"`                                          // TLS fingerprint or similar identifier.
	IsPrivate     bool                   `json:"is_private" gorm:"default:false"`                                // Specifies if the host is private; defaults to false.
	IsOnline      bool                   `json:"is_online" gorm:"default:false;index"`                           // Indicates if the host is currently online; defaults to false.
	IsFreeTier    bool                   `json:"is_free_tier" gorm:"default:false;index"`                        // Specifies if the host is available for the free tier; defaults to false.
	Status        customTypes.HostStatus `json:"status,omitempty" gorm:"type:varchar(20);default:'unknown'"`     // Detailed status of the host (e.g., active, maintenance); defaults to 'unknown'.
	LastCheckedAt *time.Time             `json:"last_checked_at,omitempty"`                                      // Timestamp of the last status check.
	CreatedAt     time.Time              `json:"created_at"`                                                     // Timestamp of creation.
	UpdatedAt     time.Time              `json:"updated_at"`                                                     // Timestamp of the last update.
	DeletedAt     gorm.DeletedAt         `gorm:"index" json:"deleted_at,omitempty"`                              // Timestamp for soft deletion.
}

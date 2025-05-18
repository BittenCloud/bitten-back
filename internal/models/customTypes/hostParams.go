package customTypes

// ListHostsParams contains parameters for filtering and paginating the list of hosts.
// Pointer fields are used for optional filters; if a field is nil, the filter is not applied.
type ListHostsParams struct {
	Offset    int         // The number of records to skip for pagination.
	Limit     int         // The maximum number of records to return.
	Country   *string     // Optional: Filter by country code (e.g., ISO 3166-1 alpha-2).
	City      *string     // Optional: Filter by city name.
	Protocol  *string     // Optional: Filter by protocol (e.g., "tcp", "udp", "http").
	Network   *string     // Optional: Filter by network type (e.g., "tcp", "ws").
	IsOnline  *bool       // Optional: Filter by online status.
	IsPrivate *bool       // Optional: Filter by private status.
	Status    *HostStatus // Optional: Filter by specific host status (e.g., "active", "maintenance").
	HostName  *string     // Optional: Filter by a partial match on the host name.
	Address   *string     // Optional: Filter by a partial match on the host address (IP or domain).
	SortBy    string      // Field name to sort by (e.g., "created_at", "host_name").
	SortOrder string      // Sort order: "asc" for ascending, "desc" for descending.
}

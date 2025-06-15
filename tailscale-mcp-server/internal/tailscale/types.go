package tailscale

import "time"

// Device represents a Tailscale device
type Device struct {
	ID                        string              `json:"id"`
	Name                      string              `json:"name"`
	Hostname                  string              `json:"hostname"`
	ClientVersion             string              `json:"clientVersion"`
	UpdateAvailable           bool                `json:"updateAvailable"`
	OS                        string              `json:"os"`
	Created                   time.Time           `json:"created,omitempty"`
	LastSeen                  time.Time           `json:"lastSeen"`
	KeyExpiryDisabled         bool                `json:"keyExpiryDisabled"`
	Expires                   time.Time           `json:"expires"`
	Authorized                bool                `json:"authorized"`
	IsExternal                bool                `json:"isExternal"`
	MachineKey                string              `json:"machineKey"`
	NodeKey                   string              `json:"nodeKey"`
	BlocksIncomingConnections bool                `json:"blocksIncomingConnections"`
	EnabledRoutes             []string            `json:"enabledRoutes"`
	AdvertisedRoutes          []string            `json:"advertisedRoutes"`
	ClientConnectivity        *ClientConnectivity `json:"clientConnectivity,omitempty"`
	Addresses                 []string            `json:"addresses"`
	Tags                      []string            `json:"tags"`
	TailnetLockError          string              `json:"tailnetLockError,omitempty"`
	TailnetLockKey            string              `json:"tailnetLockKey,omitempty"`
	User                      string              `json:"user"`
}

// ClientConnectivity represents device connectivity information
type ClientConnectivity struct {
	Endpoints             []string           `json:"endpoints"`
	Derp                  string             `json:"derp"`
	MappingVariesByDestIP bool               `json:"mappingVariesByDestIP"`
	Latency               map[string]Latency `json:"latency"`
	ClientSupports        ClientSupports     `json:"clientSupports"`
}

// Latency represents latency information
type Latency struct {
	PreferredDERP int                `json:"preferredDERP"`
	DERPLatency   map[string]float64 `json:"derpLatency"`
}

// ClientSupports represents client capabilities
type ClientSupports struct {
	HairPinning bool `json:"hairPinning"`
	IPv6        bool `json:"ipv6"`
	PCP         bool `json:"pcp"`
	PMP         bool `json:"pmp"`
	UPnP        bool `json:"upnp"`
	UDP         bool `json:"udp"`
}

// DeviceRoutes represents device route information
type DeviceRoutes struct {
	AdvertisedRoutes []string `json:"advertisedRoutes"`
	EnabledRoutes    []string `json:"enabledRoutes"`
}

// DeviceAuthorization represents device authorization request
type DeviceAuthorization struct {
	Authorized bool `json:"authorized"`
}

// DeviceKey represents device key information
type DeviceKey struct {
	KeyExpiryDisabled bool `json:"keyExpiryDisabled"`
}

// DeviceTags represents device tags
type DeviceTags struct {
	Tags []string `json:"tags"`
}

// TailnetInfo represents tailnet information
type TailnetInfo struct {
	Name      string    `json:"name"`
	AccountID string    `json:"accountId"`
	CreatedAt time.Time `json:"createdAt"`
	DNSConfig DNSConfig `json:"dnsConfig"`
}

// DNSConfig represents DNS configuration
type DNSConfig struct {
	Nameservers []string `json:"nameservers"`
	SearchPaths []string `json:"searchPaths"`
	MagicDNS    bool     `json:"magicDNS"`
}

// AuthKey represents an authentication key
type AuthKey struct {
	ID           string              `json:"id"`
	Key          string              `json:"key"`
	Description  string              `json:"description"`
	Created      time.Time           `json:"created"`
	Expires      time.Time           `json:"expires"`
	Revoked      bool                `json:"revoked"`
	Capabilities AuthKeyCapabilities `json:"capabilities"`
}

// AuthKeyCapabilities represents auth key capabilities
type AuthKeyCapabilities struct {
	Devices AuthKeyDeviceCapabilities `json:"devices"`
}

// AuthKeyDeviceCapabilities represents device-specific auth key capabilities
type AuthKeyDeviceCapabilities struct {
	Create AuthKeyDeviceCreateCapabilities `json:"create"`
}

// AuthKeyDeviceCreateCapabilities represents device creation capabilities
type AuthKeyDeviceCreateCapabilities struct {
	Reusable      bool     `json:"reusable"`
	Ephemeral     bool     `json:"ephemeral"`
	Preauthorized bool     `json:"preauthorized"`
	Tags          []string `json:"tags"`
}

// AuthKeyRequest represents a request to create an auth key
type AuthKeyRequest struct {
	Capabilities  AuthKeyCapabilities `json:"capabilities"`
	ExpirySeconds int                 `json:"expirySeconds,omitempty"`
	Description   string              `json:"description,omitempty"`
}

// DeviceListResponse represents the response from listing devices
type DeviceListResponse struct {
	Devices []Device `json:"devices"`
}

// AuthKeyListResponse represents the response from listing auth keys
type AuthKeyListResponse struct {
	Keys []AuthKey `json:"keys"`
}

////////////////////////////////////////////////////////////////////////////////
// CLI types
////////////////////////////////////////////////////////////////////////////////

type CLIResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Stderr  string `json:"stderr,omitempty"`
	Error   string `json:"error,omitempty"`
}

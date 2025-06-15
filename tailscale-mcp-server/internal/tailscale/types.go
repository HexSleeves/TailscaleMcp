// tailscale-mcp-server/internal/tailscale/types.go
package tailscale

import (
	"fmt"
	"time"
)

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

// String returns a human-readable representation of the device
func (d Device) String() string {
	status := "unauthorized"
	if d.Authorized {
		status = "authorized"
	}
	return fmt.Sprintf("Device{Name: %s, ID: %s, Status: %s}", d.Name, d.ID, status)
}

// IsOnline returns true if the device was seen recently (within last 5 minutes)
func (d Device) IsOnline() bool {
	return time.Since(d.LastSeen) < 5*time.Minute
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

// String returns a human-readable representation of the tailnet
func (t TailnetInfo) String() string {
	return fmt.Sprintf("Tailnet{Name: %s, AccountID: %s}", t.Name, t.AccountID)
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

// IsExpired returns true if the auth key has expired
func (a AuthKey) IsExpired() bool {
	return time.Now().After(a.Expires)
}

// IsValid returns true if the auth key is not revoked and not expired
func (a AuthKey) IsValid() bool {
	return !a.Revoked && !a.IsExpired()
}

// String returns a human-readable representation of the auth key
func (a AuthKey) String() string {
	status := "valid"
	if a.Revoked {
		status = "revoked"
	} else if a.IsExpired() {
		status = "expired"
	}
	return fmt.Sprintf("AuthKey{ID: %s, Description: %s, Status: %s}", a.ID, a.Description, status)
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

// Count returns the number of devices
func (d DeviceListResponse) Count() int {
	return len(d.Devices)
}

// AuthorizedDevices returns only authorized devices
func (d DeviceListResponse) AuthorizedDevices() []Device {
	var authorized []Device
	for _, device := range d.Devices {
		if device.Authorized {
			authorized = append(authorized, device)
		}
	}
	return authorized
}

// OnlineDevices returns only devices that appear to be online
func (d DeviceListResponse) OnlineDevices() []Device {
	var online []Device
	for _, device := range d.Devices {
		if device.IsOnline() {
			online = append(online, device)
		}
	}
	return online
}

// AuthKeyListResponse represents the response from listing auth keys
type AuthKeyListResponse struct {
	Keys []AuthKey `json:"keys"`
}

// ValidKeys returns only valid (not revoked, not expired) auth keys
func (a AuthKeyListResponse) ValidKeys() []AuthKey {
	var valid []AuthKey
	for _, key := range a.Keys {
		if key.IsValid() {
			valid = append(valid, key)
		}
	}
	return valid
}

////////////////////////////////////////////////////////////////////////////////
// Error types
////////////////////////////////////////////////////////////////////////////////

// APIError represents a Tailscale API error with additional context
type APIError struct {
	Operation  string `json:"operation"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	RequestID  string `json:"requestId,omitempty"`
}

func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("Tailscale API error in %s (status %d, request %s): %s",
			e.Operation, e.StatusCode, e.RequestID, e.Message)
	}
	return fmt.Sprintf("Tailscale API error in %s (status %d): %s",
		e.Operation, e.StatusCode, e.Message)
}

// IsNotFound returns true if this is a 404 error
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsUnauthorized returns true if this is a 401 error
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == 401
}

// IsForbidden returns true if this is a 403 error
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == 403
}

// IsRateLimited returns true if this is a 429 error
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == 429
}

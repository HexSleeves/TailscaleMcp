package cli

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
)

// Status corresponds to the top-level object returned by
//
//	tailscale status --json
type TailscaleStatus struct {
	Version        string          `json:"Version" validate:"required"`
	TUN            bool            `json:"TUN"`
	BackendState   string          `json:"BackendState" validate:"required"`
	HaveNodeKey    bool            `json:"HaveNodeKey"`
	AuthURL        string          `json:"AuthURL"`
	TailscaleIPs   []string        `json:"TailscaleIPs" validate:"dive,ip"`
	Self           SelfStatus      `json:"Self"`
	Health         []string        `json:"Health,omitempty"`
	MagicDNSSuffix string          `json:"MagicDNSSuffix,omitempty"`
	CurrentTailnet *CurrentTailnet `json:"CurrentTailnet,omitempty"`
	CertDomains    []string        `json:"CertDomains,omitempty" validate:"dive,fqdn"`
	Peer           map[string]Peer `json:"Peer,omitempty"`
	User           map[string]User `json:"User,omitempty"`
	ClientVersion  *ClientVersion  `json:"ClientVersion,omitempty"`
}

// -----------------------------------------------------------------------------
// Embedded objects
// -----------------------------------------------------------------------------

type SelfStatus struct {
	ID                  string         `json:"ID"`
	PublicKey           string         `json:"PublicKey"`
	HostName            string         `json:"HostName"`
	DNSName             string         `json:"DNSName" validate:"omitempty,fqdn"`
	OS                  string         `json:"OS"`
	UserID              int64          `json:"UserID" validate:"omitempty,min=1"`
	TailscaleIPs        []string       `json:"TailscaleIPs" validate:"dive,ip"`
	AllowedIPs          []string       `json:"AllowedIPs,omitempty" validate:"dive,cidr"`
	Addrs               []string       `json:"Addrs,omitempty"`
	CurAddr             string         `json:"CurAddr"`
	Relay               string         `json:"Relay"`
	RxBytes             uint64         `json:"RxBytes"`
	TxBytes             uint64         `json:"TxBytes"`
	Created             string         `json:"Created"`
	LastWrite           string         `json:"LastWrite"`
	LastSeen            string         `json:"LastSeen"`
	LastHandshake       string         `json:"LastHandshake"`
	Online              bool           `json:"Online"`
	ExitNode            bool           `json:"ExitNode"`
	ExitNodeOption      bool           `json:"ExitNodeOption"`
	Active              bool           `json:"Active"`
	PeerAPIURL          []string       `json:"PeerAPIURL,omitempty" validate:"dive,url"`
	TaildropTarget      int            `json:"TaildropTarget"`
	NoFileSharingReason string         `json:"NoFileSharingReason"`
	Capabilities        []string       `json:"Capabilities,omitempty"`
	CapMap              map[string]any `json:"CapMap,omitempty"`
	InNetworkMap        bool           `json:"InNetworkMap"`
	InMagicSock         bool           `json:"InMagicSock"`
	InEngine            bool           `json:"InEngine"`
	KeyExpiry           string         `json:"KeyExpiry"`
}

// Peer re-uses almost everything from SelfStatus and adds a few extras.
type Peer struct {
	SelfStatus             // embed → promotes identical fields
	PrimaryRoutes []string `json:"PrimaryRoutes,omitempty" validate:"dive,cidr"`
	Expired       bool     `json:"Expired,omitempty"`
	SSHHostKeys   []string `json:"sshHostKeys,omitempty"`
}

// CurrentTailnet holds metadata about the tailnet we belong to.
type CurrentTailnet struct {
	Name            string `json:"Name" validate:"required"`
	MagicDNSSuffix  string `json:"MagicDNSSuffix" validate:"required"`
	MagicDNSEnabled bool   `json:"MagicDNSEnabled"`
}

// User maps the numeric UserID to account information.
type User struct {
	ID            int64  `json:"ID" validate:"required,min=1"`
	LoginName     string `json:"LoginName" validate:"required"`
	DisplayName   string `json:"DisplayName" validate:"required"`
	ProfilePicURL string `json:"ProfilePicURL" validate:"omitempty,url"`
}

// ClientVersion tells whether the local client is current.
type ClientVersion struct {
	RunningLatest bool `json:"RunningLatest"`
}

type Validator[T any] func(T) error

var validate = validator.New()

// ParseSchema unmarshals a JSON string into any Go type.
// If unmarshalling fails it returns the zero value of T and an error.
func ParseSchema[T any](raw string) (T, error) {
	return ParseSchemaWithValidator[T](raw, nil)
}

func ParseSchemaWithValidator[T any](
	raw string,
	customValidator Validator[T], // may be nil
) (T, error) {
	var dst T

	if err := json.Unmarshal([]byte(raw), &dst); err != nil {
		var zero T
		return zero, err
	}

	// Use validator package for struct validation
	if err := validate.Struct(dst); err != nil {
		var zero T
		return zero, err
	}

	// Apply custom validator if provided
	if customValidator != nil {
		if err := customValidator(dst); err != nil {
			var zero T
			return zero, err
		}
	}

	return dst, nil
}

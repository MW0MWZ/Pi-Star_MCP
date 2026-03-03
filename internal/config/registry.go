package config

import "sort"

// ServiceCategory classifies a service for grouping in the UI.
type ServiceCategory int

const (
	CategoryCore    ServiceCategory = iota
	CategoryGateway
	CategoryBridge
	CategoryUtility
)

// ArgStyle describes how a service binary receives its arguments.
type ArgStyle int

const (
	// ArgPositional: binary configpath  (e.g. MMDVMHost /etc/mmdvmhost/MMDVM.ini)
	ArgPositional ArgStyle = iota
	// ArgFlag: binary -config configpath  (e.g. dstargateway -config /etc/...)
	ArgFlag
	// ArgPort: binary port  (e.g. YSFParrot 4001) — no config file
	ArgPort
)

// ServiceDef is a compile-time definition of a managed service.
type ServiceDef struct {
	Name              string
	DisplayName       string
	Category          ServiceCategory
	DefaultBinaryPath string
	DefaultConfigPath string // empty for ArgPort services
	DefaultArgs       string // for ArgPort services: default port number
	ConfigArgStyle    ArgStyle
	DependsOn         []string // service names that must be running first
	DefaultEnabled    bool
}

// Registry maps service name → definition for every supported service.
var Registry = map[string]ServiceDef{
	// ── Core ──────────────────────────────────────────────
	"mmdvmhost": {
		Name:              "mmdvmhost",
		DisplayName:       "MMDVMHost",
		Category:          CategoryCore,
		DefaultBinaryPath: "/usr/local/bin/MMDVMHost",
		DefaultConfigPath: "/etc/mmdvmhost/MMDVM.ini",
		ConfigArgStyle:    ArgPositional,
		DefaultEnabled:    true,
	},

	// ── Gateways ─────────────────────────────────────────
	"dmrgateway": {
		Name:              "dmrgateway",
		DisplayName:       "DMRGateway",
		Category:          CategoryGateway,
		DefaultBinaryPath: "/usr/local/bin/DMRGateway",
		DefaultConfigPath: "/etc/dmrclients/DMRGateway.ini",
		ConfigArgStyle:    ArgPositional,
	},
	"ysfgateway": {
		Name:              "ysfgateway",
		DisplayName:       "YSFGateway",
		Category:          CategoryGateway,
		DefaultBinaryPath: "/usr/local/bin/YSFGateway",
		DefaultConfigPath: "/etc/ysfclients/YSFGateway.ini",
		ConfigArgStyle:    ArgPositional,
	},
	"p25gateway": {
		Name:              "p25gateway",
		DisplayName:       "P25Gateway",
		Category:          CategoryGateway,
		DefaultBinaryPath: "/usr/local/bin/P25Gateway",
		DefaultConfigPath: "/etc/p25clients/P25Gateway.ini",
		ConfigArgStyle:    ArgPositional,
	},
	"nxdngateway": {
		Name:              "nxdngateway",
		DisplayName:       "NXDNGateway",
		Category:          CategoryGateway,
		DefaultBinaryPath: "/usr/local/bin/NXDNGateway",
		DefaultConfigPath: "/etc/nxdnclients/NXDNGateway.ini",
		ConfigArgStyle:    ArgPositional,
	},
	"dstargateway": {
		Name:              "dstargateway",
		DisplayName:       "dstargateway",
		Category:          CategoryGateway,
		DefaultBinaryPath: "/usr/local/bin/dstargateway",
		DefaultConfigPath: "/etc/dstarclients/dstargateway.cfg",
		ConfigArgStyle:    ArgFlag,
	},
	"fmgateway": {
		Name:              "fmgateway",
		DisplayName:       "FMGateway",
		Category:          CategoryGateway,
		DefaultBinaryPath: "/usr/local/bin/FMGateway",
		DefaultConfigPath: "/etc/fmclients/FMGateway.ini",
		ConfigArgStyle:    ArgPositional,
		DependsOn:         []string{"mmdvmhost"},
	},
	"aprsgateway": {
		Name:              "aprsgateway",
		DisplayName:       "APRSGateway",
		Category:          CategoryGateway,
		DefaultBinaryPath: "/usr/local/bin/APRSGateway",
		DefaultConfigPath: "/etc/aprsclients/APRSGateway.ini",
		ConfigArgStyle:    ArgPositional,
		DependsOn:         []string{"mmdvmhost"},
	},
	"dapnetgateway": {
		Name:              "dapnetgateway",
		DisplayName:       "DAPNETGateway",
		Category:          CategoryGateway,
		DefaultBinaryPath: "/usr/local/bin/DAPNETGateway",
		DefaultConfigPath: "/etc/pocsagclients/DAPNETGateway.ini",
		ConfigArgStyle:    ArgPositional,
		DependsOn:         []string{"mmdvmhost"},
	},

	// ── Bridges ──────────────────────────────────────────
	"dmr2ysf": {
		Name:              "dmr2ysf",
		DisplayName:       "DMR2YSF",
		Category:          CategoryBridge,
		DefaultBinaryPath: "/usr/local/bin/DMR2YSF",
		DefaultConfigPath: "/etc/dmrclients/DMR2YSF.ini",
		ConfigArgStyle:    ArgPositional,
		DependsOn:         []string{"dmrgateway"},
	},
	"dmr2nxdn": {
		Name:              "dmr2nxdn",
		DisplayName:       "DMR2NXDN",
		Category:          CategoryBridge,
		DefaultBinaryPath: "/usr/local/bin/DMR2NXDN",
		DefaultConfigPath: "/etc/dmrclients/DMR2NXDN.ini",
		ConfigArgStyle:    ArgPositional,
		DependsOn:         []string{"dmrgateway"},
	},
	"ysf2dmr": {
		Name:              "ysf2dmr",
		DisplayName:       "YSF2DMR",
		Category:          CategoryBridge,
		DefaultBinaryPath: "/usr/local/bin/YSF2DMR",
		DefaultConfigPath: "/etc/ysfclients/YSF2DMR.ini",
		ConfigArgStyle:    ArgPositional,
	},
	"ysf2nxdn": {
		Name:              "ysf2nxdn",
		DisplayName:       "YSF2NXDN",
		Category:          CategoryBridge,
		DefaultBinaryPath: "/usr/local/bin/YSF2NXDN",
		DefaultConfigPath: "/etc/ysfclients/YSF2NXDN.ini",
		ConfigArgStyle:    ArgPositional,
	},
	"ysf2p25": {
		Name:              "ysf2p25",
		DisplayName:       "YSF2P25",
		Category:          CategoryBridge,
		DefaultBinaryPath: "/usr/local/bin/YSF2P25",
		DefaultConfigPath: "/etc/ysfclients/YSF2P25.ini",
		ConfigArgStyle:    ArgPositional,
	},
	"nxdn2dmr": {
		Name:              "nxdn2dmr",
		DisplayName:       "NXDN2DMR",
		Category:          CategoryBridge,
		DefaultBinaryPath: "/usr/local/bin/NXDN2DMR",
		DefaultConfigPath: "/etc/nxdnclients/NXDN2DMR.ini",
		ConfigArgStyle:    ArgPositional,
		DependsOn:         []string{"nxdngateway"},
	},

	// ── Utilities ────────────────────────────────────────
	"ysfparrot": {
		Name:              "ysfparrot",
		DisplayName:       "YSFParrot",
		Category:          CategoryUtility,
		DefaultBinaryPath: "/usr/local/bin/YSFParrot",
		DefaultArgs:       "42014",
		ConfigArgStyle:    ArgPort,
	},
	"p25parrot": {
		Name:              "p25parrot",
		DisplayName:       "P25Parrot",
		Category:          CategoryUtility,
		DefaultBinaryPath: "/usr/local/bin/P25Parrot",
		DefaultArgs:       "42015",
		ConfigArgStyle:    ArgPort,
	},
	"nxdnparrot": {
		Name:              "nxdnparrot",
		DisplayName:       "NXDNParrot",
		Category:          CategoryUtility,
		DefaultBinaryPath: "/usr/local/bin/NXDNParrot",
		DefaultArgs:       "42016",
		ConfigArgStyle:    ArgPort,
	},
	"dstarrepeater": {
		Name:              "dstarrepeater",
		DisplayName:       "DStarRepeater",
		Category:          CategoryCore,
		DefaultBinaryPath: "/usr/bin/dstarrepeaterd",
		DefaultConfigPath: "/etc/dstarrepeater/dstarrepeater.conf",
		ConfigArgStyle:    ArgFlag,
	},
	"dgidgateway": {
		Name:              "dgidgateway",
		DisplayName:       "DGIdGateway",
		Category:          CategoryUtility,
		DefaultBinaryPath: "/usr/local/bin/DGIdGateway",
		DefaultConfigPath: "/etc/ysfclients/DGIdGateway.ini",
		ConfigArgStyle:    ArgPositional,
	},
}

// DStarVariant maps a DStarRepeater hardware type to its binary and
// config file. Each hardware type requires a different daemon binary.
type DStarVariant struct {
	Key         string // stored in dashboard.ini
	DisplayName string
	BinaryName  string // filename in /usr/bin/ (e.g. "dstarrepeaterd")
	ConfigFile  string // filename in /etc/dstarrepeater/ (e.g. "dstarrepeater.conf")
}

// DStarVariants lists all supported DStarRepeater hardware types.
// The first entry is the default.
var DStarVariants = []DStarVariant{
	{Key: "icom", DisplayName: "Icom Repeater (ID-RP2C)", BinaryName: "dstarrepeaterd", ConfigFile: "dstarrepeater.conf"},
	{Key: "dvmega", DisplayName: "DV-Mega", BinaryName: "dstarrepeaterd", ConfigFile: "dstarrepeater.conf"},
	{Key: "gmsk", DisplayName: "GMSK Modem", BinaryName: "gmskrepeaterd", ConfigFile: "gmskrepeater.conf"},
	{Key: "dvrptr1", DisplayName: "DV-RPTR V1", BinaryName: "dvrptrrepeaterd", ConfigFile: "dvrptrrepeater.conf"},
	{Key: "dvrptr3", DisplayName: "DV-RPTR V3", BinaryName: "dvrptrrepeaterd", ConfigFile: "dvrptrrepeater.conf"},
	{Key: "dvap", DisplayName: "DVAP", BinaryName: "dvapnoded", ConfigFile: "dvapnode.conf"},
}

// LookupDStarVariant returns the variant for a hardware type key.
// Returns the default (first) variant if the key is unknown.
func LookupDStarVariant(key string) DStarVariant {
	for _, v := range DStarVariants {
		if v.Key == key {
			return v
		}
	}
	return DStarVariants[0]
}

// LookupService returns the definition for a named service.
func LookupService(name string) (ServiceDef, bool) {
	def, ok := Registry[name]
	return def, ok
}

// ServiceNames returns all registered service names in sorted order.
func ServiceNames() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

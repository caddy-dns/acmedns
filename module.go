package acmedns

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	libdnsacmedns "github.com/libdns/acmedns"
)

// Provider lets Caddy read and manipulate DNS records hosted by this DNS provider.
type Provider struct {
	libdnsacmedns.Provider
	ConfigFilePath string `json:"config_file_path,omitempty"`
}

func init() {
	caddy.RegisterModule(Provider{})
}

// CaddyModule returns the Caddy module information.
func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dns.providers.acmedns",
		New: func() caddy.Module { return &Provider{} },
	}
}

// Provision sets up the module. Implements caddy.Provisioner.
func (p *Provider) Provision(ctx caddy.Context) error {
	file, err := ioutil.ReadFile(p.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("Failed to read config file %s", p.ConfigFilePath)
	}
	var configs map[string]libdnsacmedns.DomainConfig
	err = json.Unmarshal(file, &configs)
	p.Configs = configs
	if err != nil {
		return fmt.Errorf("Failed to unmarshall config")
	}
	return nil
}

// UnmarshalCaddyfile sets up the DNS provider from Caddyfile tokens. Syntax:
//
//
// acmedns <config_file_path>
//
// or
//
// acmedns {
//     config_file_path <config_file_path>
// }
func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			p.ConfigFilePath = d.Val()
		}
		if d.NextArg() {
			return d.ArgErr()
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "config_file_path":
				if p.ConfigFilePath != "" {
					return d.Err("config_file_path already set")
				}
				if d.NextArg() {
					p.ConfigFilePath = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
		}
	}
	if p.ConfigFilePath == "" {
		return d.Err("missing config_file_path")
	}
	return nil
}

// Interface guards
var (
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ caddy.Provisioner     = (*Provider)(nil)
)

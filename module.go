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
	if p.ConfigFilePath == "" {
		return nil
	}
	file, err := ioutil.ReadFile(p.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("Failed to read config file %s", p.ConfigFilePath)
	}
	var configs map[string]libdnsacmedns.DomainConfig
	err = json.Unmarshal(file, &configs)
	if err == nil {
		p.Configs = configs
		return nil
	}

	// err is not nil, trying a different config format
	var config libdnsacmedns.Provider
	err = json.Unmarshal(file, &config)
	if err == nil {
		p.Username = config.Username
		p.Password = config.Password
		p.Subdomain = config.Subdomain
		p.ServerURL = config.ServerURL
		return nil
	}

	return fmt.Errorf("Failed to unmarshall config")
}

// UnmarshalCaddyfile sets up the DNS provider from Caddyfile tokens.
// There are four alternative ways to configure acmedns provider.
//
// 1)
// acmedns <config_file_path>
//
// 2)
// acmedns {
//     config_file_path <config_file_path>
// }
//
// 3)
// acmedns {
//	   username <username>
//     password <password>
//     subdomain <subdomain>
//     server_url <server_url>
// }
//
// 4)
// acmedns {
//     config {
//         domain1.example.com {
//	           username <username>
//             password <password>
//             subdomain <subdomain>
//             fulldomain <fulldomain>
//             server_url <server_url>
//         }
//         domain2.example.com {
//	           username <username>
//             password <password>
//             subdomain <subdomain>
//             fulldomain <fulldomain>
//             server_url <server_url>
//         }
//     }
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
			case "username":
				if p.Username != "" {
					return d.Err("username already set")
				}
				if d.NextArg() {
					p.Username = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "password":
				if p.Password != "" {
					return d.Err("password already set")
				}
				if d.NextArg() {
					p.Password = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "subdomain":
				if p.Subdomain != "" {
					return d.Err("subdomain already set")
				}
				if d.NextArg() {
					p.Subdomain = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "server_url":
				if p.ServerURL != "" {
					return d.Err("server_url already set")
				}
				if d.NextArg() {
					p.ServerURL = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "config":
				configs, err := unmarshallConfigs(d)
				if err != nil {
					return err
				}
				p.Configs = configs
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
		}
	}
	return p.verifyUnmarshalled(d)
}

func (p *Provider) verifyUnmarshalled(d *caddyfile.Dispenser) error {
	useOneAccountConfig :=
		p.Username != "" &&
			p.Password != "" &&
			p.Subdomain != "" &&
			p.ServerURL != ""
	useFileConfig := p.ConfigFilePath != ""
	useMultiAccountConfig := p.Configs != nil
	useCount := 0
	if useOneAccountConfig {
		useCount += 1
	}
	if useFileConfig {
		useCount += 1
	}
	if useMultiAccountConfig {
		useCount += 1
	}
	if useCount != 1 {
		return fmt.Errorf("Failed to parse acmedns configuration. " +
			"You must use exactly one of these directives/directive combinations: " +
			"1) config, 2) config_file_path, " +
			"3) username, password, subdomain and server_url combination")
	}

	if useMultiAccountConfig && len(p.Configs) == 0 {
		return d.Errf("config section must have at least one domain")
	}

	return nil
}

func unmarshallConfigs(d *caddyfile.Dispenser) (map[string]libdnsacmedns.DomainConfig, error) {
	configs := make(map[string]libdnsacmedns.DomainConfig)
	for configNesting := d.Nesting(); d.NextBlock(configNesting); {
		domainConfig := libdnsacmedns.DomainConfig{}
		domain := d.Val()
		for domainNesting := d.Nesting(); d.NextBlock(domainNesting); {
			switch d.Val() {
			case "username":
				if d.NextArg() {
					domainConfig.Username = d.Val()
				}
				if d.NextArg() {
					return nil, d.ArgErr()
				}
			case "password":
				if d.NextArg() {
					domainConfig.Password = d.Val()
				}
				if d.NextArg() {
					return nil, d.ArgErr()
				}
			case "subdomain":
				if d.NextArg() {
					domainConfig.Subdomain = d.Val()
				}
				if d.NextArg() {
					return nil, d.ArgErr()
				}
			case "fulldomain":
				if d.NextArg() {
					domainConfig.FullDomain = d.Val()
				}
				if d.NextArg() {
					return nil, d.ArgErr()
				}
			case "server_url":
				if d.NextArg() {
					domainConfig.ServerURL = d.Val()
				}
				if d.NextArg() {
					return nil, d.ArgErr()
				}
			default:
				return nil, d.Errf("unrecognized subdirective: '%s'", d.Val())
			}
		}
		configs[domain] = domainConfig
	}
	return configs, nil
}

// Interface guards
var (
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ caddy.Provisioner     = (*Provider)(nil)
)

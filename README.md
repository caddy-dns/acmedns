ACME-DNS module for Caddy
===========================

This package contains a DNS provider module for [Caddy](https://github.com/caddyserver/caddy). It can be used to solve [DNS-01 challenges](https://letsencrypt.org/docs/challenge-types/) with [ACME-DNS](https://github.com/joohoi/acme-dns) server.

## Caddy module name

```
dns.providers.acmedns
```

## Using ACME-DNS to obtain HTTPS certificates with Caddy

To use ACME-DNS for solving [DNS-01 challenge](https://letsencrypt.org/docs/challenge-types/) and obtaining a certificate, you'll need:

* Caddy version with this plugin built-in. See [xcaddy](https://github.com/caddyserver/xcaddy) to learn how to build Caddy with plugins.
* A domain name that you control. In this example, we'll assume it's `your-domain.example.com`. You'll need to be able to create a CNAME record with name `_acme-challenge.your-domain.example.com`.
* An access to ACME-DNS server. For testing purposes, you can you the public server at [https://auth.acme-dns.io](). However, self-hosting is highly encouraged. To learn how to self-host ACME-DNS server, refer to [ACME-DNS documentation](https://github.com/joohoi/acme-dns#self-hosted).

The follow these steps:

1. Register and account on ACME-DNS server. Refer to [ACME-DNS documentation](https://github.com/joohoi/acme-dns#register-endpoint). In short: make a POST request to `<ACME-DNS server URL>/register`, i.e. run:
	`curl -X POST https://auth.acme-dns.io/register`

	The response should be a JSON that contains your new credentials, looking similar to this one:

	```
	{
		"username": "5d26a340-2e1d-4b6b-af1a-4aab569897b7",
		"password": "_r2gFOVtrF9I82_l6ZXTfPaxCgldqJSWaTmd4BS9",
		"fulldomain": "37c51280-79ca-435f-af32-c775eb67e2ab.auth.acme-dns.io",
		"subdomain": "37c51280-79ca-435f-af32-c775eb67e2ab",
		"allowfrom": []
	}
	```

2. Create a DNS CNAME record that points from `_acme-challenge.your-domain.example.com` to the `fulldomain` from the registration response. In this case, it would be

    `_acme-challenge.your-domain.example.com.   CNAME   37c51280-79ca-435f-af32-c775eb67e2ab.auth.acme-dns.io.`

3. Use the credentials obtained in step 1 to configure `acmedns` plugin in Caddy. This is a simple example of a working Caddyfile:

    ```
	your-domain.example.com
	
	tls {
		dns acmedns {
			username <username you obtained in step 1>
			password <password you obtained in steip 1>
			subdomain <ACME-DNS subdomain you obtained in step 1>
			server_url <ACME-DNS server API URL>   # e.g. https://auth.acme-dns.io
		}
	}
	
	respond "Hello"
	```

## More configuration options

There are two orthogonal choices that you can make about the configuration of `acmedns` plugin.

1. Whether to put the credentials directly in `Caddyfile` / `caddy.json` _or_ to use a separate configuration file.
2. Whether to use a simple one-account configuration as described in the previous section _or_ to use a multi-account set up (more on that below).

### JSON credentials file

You can save `acmedns` account credentials as a JSON file instead or writing it directly to `Caddyfile`/`caddy.json`. The credentials file should look like this:

```
{
	"username": "<username>",
	"password": "<password>",
	"subdomain": "37c51280-79ca-435f-af32-c775eb67e2ab",
	"server_url": "<server_url>"
}
```

And your `Caddyfile`:

```
your-domain.example.com
tls {
	dns acmedns /path/to/credentials.json
}
respond "Hello"
```

### Multi-account configuration

Simple configuration showed in the previous section is enough for most use-cases. If you are serving multiple domains, you can use different `dns acmedns {...}` directives for different site blocks and thus use different ACME-DNS accounts per domain:

```
your-domain-1.example.com {
	tls {
		dns acmedns {
			username <username 1>
			...
		}
	}
	respond "Hello 1"
}

your-domain-2.example.com {
	tls {
		dns acmedns {
			username <username 2>
			...
		}
	}
	respond "Hello 2"
}
```

However, you can also provide `acmedns` plugin with a single configuration which contains domain->account/credentials mapping. The credentials file will look like this:

```
{
    "your-domain-1.example.com": {
        "username": "<username>",
        "password": "<password>",
        "fulldomain": "<full domain from registration response",
        "subdomain": "<subdomain>",
        "server_url": "<server URL>"
    },
	"your-domain-2.example.com": {
		...
	}
}
```

Not that this type of configuration requires one more field -- `fulldomain` (as returned by registration endpoint).

You can also embed this into `Caddyfile` directly. If you want this configuration to apply to all site-blocks in your `Caddyfile`, you can use [acme_dns global option](https://caddyserver.com/docs/caddyfile/options#acme-dns).

```
{
	acme_dns acmedns {
		config {
			your-domain-1.example.com {
				username <username>
				password <password>
				subdomain <subdomain>
				server_url <server url>

				# Note that this type of configuration requires one parameter
				# more -- the fulldomain value from registration response
				fulldomain <full domain>
			}
			your-domain-2.example.com {
				username <username>
				... # same fields as above
			}
		}
	}
}
your-domain-1.example.com {
	respond "Hello 1"
}
your-domain-2.example.com {
	respond "Hello 2"
}
```

Using multi-account configuration is useful if you want to manage all your configurations with [acme-dns-client CLI tool](https://github.com/acme-dns/acme-dns-client). `acme-dns-client` helps saves obtained credentials in a JSON file at `/etc/acmedns/clientstorage.json`. This file is compatible with `acmedns` Caddy plugin, you can point to it with `dns acmedns /etc/acmedns/clientstorage.json` directive (make sure that Caddy has permissions to read that file).

### Using `caddy.json` instead of `Caddyfile`

You have all the same options for configuration if you use `caddy.json` configuration format instead of `Caddyfile`. Configure your [ACME issuer](https://caddyserver.com/docs/json/apps/tls/automation/policies/issuer/acme/) like so (single-account configuration):

```
{
	"module": "acme",
	"challenges": {
		"dns": {
			"provider": {
				"name": "acmedns",
				"username": "<username>",
				"password": "<password>",
				"subdomain": "37c51280-79ca-435f-af32-c775eb67e2ab",
				"server_url": "<server_url>"
			}
		}
	}
}
```

Or like so (multi-account-configuration):

```
{
	"module": "acme",
	"challenges": {
		"dns": {
			"provider": {
				"name": "acmedns",
				"config": {
					"your-domain-1.example.com": {
						"username": "<username>",
						"password": "<password>",
						"fulldomain": "<full domain from registration response",
						"subdomain": "<subdomain>",
						"server_url": "<server URL>"
					}
				}
			}
		}
	}
}
```

Or like so (credentials file):

```
{
	"module": "acme",
	"challenges": {
		"dns": {
			"provider": {
				"name": "acmedns",
				"config_file_path": "/file/to/credentials.json"
			}
		}
	}
}
```

## Troubleshooting

If Caddy hangs on trying to obtain a certificate and later throws a timeout error, make sure that you created a correct CNAME record:

`_acme-challenge.your-domain.example.com.  CNAME  <ACME-DNS subdomain from registration>`

You can check this with `dig _acme-challenge.your-domain.example.com.`

If this record it correct, this error might be caused by DNS resolvers caching results. Using Cloudflare or Google resolvers (1.1.1.1 and 8.8.8.8) might help:

```
your-domain.example.com
tls {
	resolvers 1.1.1.1 8.8.8.8
	dns acmedns {
		...
	}
}
```

## Resources, links

1. [A Technical Deep Dive: Securing the Automation of ACME DNS Challenge Validation](https://www.eff.org/deeplinks/2018/02/technical-deep-dive-securing-automation-acme-dns-challenge-validation)

2. [ACME-DNS](https://github.com/joohoi/acme-dns)

3. [acme-dns-client](https://github.com/acme-dns/acme-dns-client)

4. [ACME-DNS provider for `libdns`](https://github.com/libdns/acmedns) -- this `acmedns` Caddy plugin depends on `libdns/acmedns` provider.
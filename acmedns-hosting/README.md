# Self-hosting ACME-DNS

This is a short guide on how to self-host ACME-DNS and how to secure its API with Caddy using URL path rewriting. We'll use Ubuntu 20.04 VPS (virtual private server) and Docker to achieve this.

Assumptions:

1. You own a domain and you can add DNS records to its zone.
2. You can `ssh` into the root account of a fresh installation of Ubuntu 20.04 server accessible from the Internet.

__Note__: this guide will not teach you how to harden your server for production, this is a minimal working example of a ACME-DNS setup.


Let's say you control domain `example.com` and want to host your ACME-DNS server at `acmedns.example.com`. Your nameserver address will be `ns.acmedns.example.com`.

1. Create `A` record that points from `ns.acmedns.example.com` to the IP address of your server:

    ```ns.acmedns.example.com.    A    <SERVER IP ADDRESS>```

2. In order to delegate resolving `*.acmedns.example.com` DNS queries to your server, create an `NS` record:

    ```acmedns.example.com.    NS    ns.acmesnd.example.com.```

3. Connect to your Ubuntu server and install dependencies (Docker, git):

    ```
    ssh root@<SERVER IP ADDRESS>
    apt-get update && apt-get upgrade -y
    apt-get install git docker.io docker-compose -y
    ```

4. Clone this repository:

    ```
    git clone https://github.com/caddy-dns/acmedns
    cd acmedns
    git checkout acmedns-hosting
    cd acmedns-hosting
    ```

5. Copy `.env.example` to `.env` (`cp .env.example .env`) and edit the values in `.env`:

    5.1. Change `ACMEDNS_DOMAIN` to your domain, i.e. `acmedns.example.com`

    5.2. Change `IP_ADDRESS` to your server's public IP address. This is used in two ways. First, it's used to add a DNS record to `acmedns-config.cfg`. Second is more subtle. Since `systemd-resolved` (a service that provides network name resolution for local programs) uses port `:53`, trying to bind on it for all IP addresses won't work. What we want to do instead is bind it only to the external IP address. So, `IP_ADDRESS` is used by Docker to bind port `:53` only for this external IP. (This depends on the VPS not being behind NAT (Network Address Translation)).

    5.4. Change `API_TOKEN` to any token that you want to use to protect the ACME-DNS API. This token will be used as a URL path prefix for the API.

6. Inspect `up.sh`and then run it:

    ```bash up.sh```

    `up.sh` injects `ACMEDNS_DOMAIN` and `IP_ADDRESS` parameters to ACME-DNS config file. Feel free to inspect `acmedns-config.cfg` and change the configuration however you see fit manually.

    Then it simply runs the `docker-compose`.

7. You should now be able to use ACME-DNS API. Create an account:

    ```curl -X POST https://acmedns.example.com/<TOKEN>/register```

    The response should look like this:

    ```
    {
        "username": "<username>",
        "password": "<password>",
        "fulldomain": "<fulldomain>",
        "subdomain": "<subdomain>",
        "allowfrom": []
    }
    ```

    Update TXT value:

    ```curl -X POST https://acmedns.example.com/<TOKEN>/update -d '{"subdomain": "<subdomain>", "txt": "___validation_token_received_from_the_ca___"}' -H "X-Api-User: <username>" -H "X-Api-Key: <password>"```

    The response should look like this:

    ```{"txt": "___validation_token_received_from_the_ca___"}```

    See if DNS records were updated accordingly:

    ```dig -t txt <fulldomain>```

After the setup, the following files are important:
* `/root/acmedns/acmedns-hosting/acmedns-data/acme-dns.db` - this contains registered accounts and TXT records.
* `/root/acmedns/acmedns-hosting/caddy-data` and `caddy-config` - this contains Caddy data (certificates, configuration).
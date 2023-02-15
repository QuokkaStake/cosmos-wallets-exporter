# cosmos-wallets-exporter

![Latest release](https://img.shields.io/github/v/release/QuokkaStake/cosmos-wallets-exporter)
[![Actions Status](https://github.com/QuokkaStake/cosmos-wallets-exporter/workflows/test/badge.svg)](https://github.com/QuokkaStake/cosmos-wallets-exporter/actions)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FQuokkaStake%2Fcosmos-wallets-exporter.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FQuokkaStake%2Fcosmos-wallets-exporter?ref=badge_shield)

cosmos-wallets-exporter is a Prometheus scraper that fetches the wallet balances from an LCD server exposed by a fullnode.

## What can I use it for?

If you have a wallet that does transactions on an app's behalf without your interaction and will stop working correctly if it cannot broadcast transactions anymore due to zero balance and not enough tokens to pay for transaction fee (some examples: Axelar's broadcaster; Sentinel's dVPN node; ReStake's bot wallets), you can use this tool to scrape the balances to Prometheus and build alerts if a wallet balance falls under a specific threshold. 

## How can I set it up?

First, you need to download the latest release from [the releases page](https://github.com/QuokkaStake/cosmos-wallets-exporter/releases/). After that, you should unzip it, and you are ready to go:

```sh
wget <the link from the releases page>
tar xvfz <file you just downloaded>
./cosmos-wallets-exporter
```

Alternatively, you can build it from source (golang >= 1.18 is required):
```sh
git clone https://github.com/QuokkaStake/cosmos-wallets-exporter.git
cd cosmos-wallets-exporter
# Either build it (this will put the resulting binary into the current folder)...
make build
# ... or install it, which will put the resulting binary into $GOPATH/bin
make install
```

To run it detached, you need to run it as a systemd service. First, we have to copy the file to the system apps folder:

```sh
sudo cp ./cosmos-wallets-exporter /usr/bin
```

Then we need to create a systemd service for our app:

```sh
sudo nano /etc/systemd/system/cosmos-wallets-exporter.service
```

You can use this template (change the user to whatever user you want this to be executed from. It's advised to create a separate user for that instead of running it from root):

```
[Unit]
Description=Cosmos Wallets Exporter
After=network-online.target

[Service]
User=<username>
TimeoutStartSec=0
CPUWeight=95
IOWeight=95
ExecStart=cosmos-wallets-exporter --config <path to config>
Restart=always
RestartSec=2
LimitNOFILE=800000
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
```

Then we'll add this service to autostart and run it:

```sh
sudo systemctl daemon-reload # reflect the systemd file change
sudo systemctl enable cosmos-wallets-exporter # enable the scraper to run on system startup
sudo systemctl start cosmos-wallets-exporter # start it
sudo systemctl status cosmos-wallets-exporter # validate it's running
```

If you need to, you can also see the logs of the process:

```sh
sudo journalctl -u cosmos-wallets-exporter -f --output cat
```

## How can I scrape data from it?

Here's the example of the Prometheus config you can use for scraping data:

```yaml
scrape-configs:
  - job_name:       'cosmos-wallets-exporter'
    scrape_interval: 30s
    static_configs:
      - targets:
        - localhost:9550 # replace localhost with scraper IP if it's on the other host
```

Then restart Prometheus and you're good to go!

All of the metrics provided by cosmos-wallets-exporter have the `cosmos_wallets_exporter_` as a prefix, here's the list of the exposed metrics:
- `cosmos_wallets_exporter_balance` - wallet balance in tokens
- `cosmos_wallets_exporter_balance_usd` - wallet balance in USD (only native tokens are used for calculation, IBC tokens are not)
- `cosmos_wallets_exporter_denom_coefficient` - coefficient between 1 display token and 1 native token (like 1 atom = 1_000_000 atom). Defaults to 1000000.
- `cosmos_wallets_exporter_success` - 1 if a wallet balance scrape was successful, 0 if no. You can also make alert to fire if it's above 0, to get notified on failed scrapes (for example, if a remote LCD endpoint is not accessible anymore). If the scrape failed, there won't be `cosmos_wallets_exporter_balance` or `cosmos_wallets_exporter_balance_usd` for it.
- `cosmos_wallets_exporter_timings` - time it took to get a response from an LCD endpoint, in seconds.

## How can I configure it?

All configuration is done via the .toml config file, which is passed to the application via the `--config` app parameter. Check `config.example.toml` for a config reference.

## How can I contribute?

Bug reports and feature requests are always welcome! If you want to contribute, feel free to open issues or PRs.


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FQuokkaStake%2Fcosmos-wallets-exporter.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FQuokkaStake%2Fcosmos-wallets-exporter?ref=badge_large)
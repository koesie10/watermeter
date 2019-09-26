# watermeter

Connects to an inductive NPN sensor and counts the number of rises, which are exported in InfluxDB and Prometheus format.

Influenced by [this post](https://ehoco.nl/watermeter-uitlezen-in-domoticz-python-script/).

Licensed under MIT license.

## Development

```
go install github.com/koesie10/watermeter/cmd/watermeter
```

## Running

### As systemd service

```
sudo mv watermeter /usr/local/bin
sudo chmod +x /usr/local/bin/watermeter

sudo adduser --system --no-create-home --group watermeter

sudo nano /etc/systemd/system/watermeter.service
```

```
[Unit]
Description=watermeter
Wants=network-online.target
After=network-online.target
After=influxdb.service
AssertFileIsExecutable=/usr/local/bin/watermeter

[Service]
User=watermeter
Group=watermeter

PermissionsStartOnly=true

Restart=always

ExecStart=/usr/local/bin/watermeter influx --influx-database telegraf --influx-tags="house=myhouse"

[Install]
WantedBy=multi-user.target
```

```
sudo systemctl daemon-reload
sudo systemctl start watermeter
sudo systemctl status watermeter
sudo systemctl enable watermeter
```

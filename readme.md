# OLED Clock

Small pixel display code showing two time zones on a WaveShare 1.3 OLED HAT for the Raspberry Pi Zero 2 W. This HAT uses the SH1106 display controller.

<https://www.waveshare.com/wiki/1.3inch_OLED_HAT>

For restart support:

create a systemd service with `sudo systemctl enable oled-clock.service`:

```
[Unit]
Description=Oled Clock
After=network.target

[Service]
ExecStart=/home/chris/oo
WorkingDirectory=/home/chris
Restart=always
User=chris

[Install]
WantedBy=multi-user.target
```

Then:

```bash
sudo systemctl enable oled-clock.service
sudo systemctl start oled-clock.service
```

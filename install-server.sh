#!/bin/bash
set -eu

sudo ln -s "$(pwd)/run-server.sh" /usr/local/bin/raspi-updater-server
sudo ln -s "$(pwd)/raspi-updater-server.service" "/etc/systemd/system/raspi-updater-server.service"
sudo systemctl daemon-reload
sudo systemctl enable raspi-updater-server.service
sudo systemctl start raspi-updater-server.service
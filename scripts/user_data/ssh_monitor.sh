#!/bin/bash

cat << 'EOF' > /usr/local/bin/ssh_monitor.sh
#!/bin/bash

log_file="/home/ec2-user/active.log"
no_ssh_count=0

while true; do
  active_users=$(who | grep -c 'pts/')
  if [ "$active_users" -eq 0 ]; then
    ((no_ssh_count++))
    if [ "$no_ssh_count" -ge 10 ]; then
      echo "$(date): No SSH sessions for 10 seconds" >> "$log_file"
      no_ssh_count=0
    fi
  else
    no_ssh_count=0
  fi
  sleep 1
done
EOF

chmod +x /usr/local/bin/ssh_monitor.sh

cat << 'EOF' > /etc/systemd/system/ssh-monitor.service
[Unit]
Description=SSH Session Monitor

[Service]
ExecStart=/usr/local/bin/ssh_monitor.sh
Restart=always
User=ec2-user

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reexec
systemctl daemon-reload
systemctl enable ssh-monitor.service
systemctl start ssh-monitor.service 
#!/bin/bash

cat << "EOF" > /usr/local/bin/ssh_monitor.sh
#!/bin/bash

# Get instance metadata
REGION=$(curl -s http://169.254.169.254/latest/meta-data/placement/region)

log_file="/home/ec2-user/active.log"
no_ssh_count=0

# Function to release lock with retries
release_lock() {
    local lock_id=$1
    local max_retries=3
    local retry_interval=2
    local retry_count=0

    while [ $retry_count -lt $max_retries ]; do
        echo "$(date): Attempting to release lock for profile $PROFILE (attempt $((retry_count + 1))/$max_retries)..." >> "$log_file"
        
        aws dynamodb delete-item \
            --region $REGION \
            --table-name dumie-lock-table \
            --key '{"LockID": {"S": "'$lock_id'"}}'

        if [ $? -eq 0 ]; then
            echo "$(date): Successfully released lock for profile $PROFILE" >> "$log_file"
            return 0
        fi

        retry_count=$((retry_count + 1))
        if [ $retry_count -lt $max_retries ]; then
            echo "$(date): Failed to release lock, retrying in ${retry_interval}s..." >> "$log_file"
            sleep $retry_interval
        fi
    done

    echo "$(date): ERROR: Failed to release lock after $max_retries attempts" >> "$log_file"
    return 1
}

while true; do
  active_users=$(who | grep -c 'pts/')
  if [ "$active_users" -eq 0 ]; then
    ((no_ssh_count++))
    if [ "$no_ssh_count" -ge 60 ]; then  # 1 minute (60 seconds)
      echo "$(date): No SSH sessions for 1 minute. Creating AMI and snapshot before termination..." >> "$log_file"
      
      # Get current instance ID when starting termination
      INSTANCE_ID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
      echo "$(date): Current instance ID: $INSTANCE_ID" >> "$log_file"
      
      # Get profile name from instance tags
      PROFILE=$(aws ec2 describe-instances \
        --region $REGION \
        --instance-ids $INSTANCE_ID \
        --query 'Reservations[0].Instances[0].Tags[?Key==`Name`].Value' \
        --output text)

      if [ -z "$PROFILE" ]; then
        echo "$(date): ERROR: Failed to get profile name from instance tags. Instance ID: $INSTANCE_ID" >> "$log_file"
        # Dump instance tags for debugging
        echo "$(date): Dumping instance tags for debugging:" >> "$log_file"
        aws ec2 describe-instances \
          --region $REGION \
          --instance-ids $INSTANCE_ID \
          --query 'Reservations[0].Instances[0].Tags' \
          --output json >> "$log_file"
        exit 1
      fi

      echo "$(date): Retrieved profile name: $PROFILE" >> "$log_file"

      # Acquire lock for this profile
      LOCK_ID="profile-$PROFILE"
      echo "$(date): Attempting to acquire lock for profile $PROFILE..." >> "$log_file"
      
      aws dynamodb put-item \
        --region $REGION \
        --table-name dumie-lock-table \
        --item '{
          "LockID": {"S": "'$LOCK_ID'"},
          "Expires": {"N": "'$(($(date +%s) + 300))'"}
        }' \
        --condition-expression "attribute_not_exists(LockID) OR Expires < :now" \
        --expression-attribute-values '{":now": {"N": "'$(date +%s)'"}}'

      if [ $? -ne 0 ]; then
        echo "$(date): Failed to acquire lock for profile $PROFILE. Another process might be using this profile." >> "$log_file"
        exit 1
      fi

      echo "$(date): Successfully acquired lock for profile $PROFILE" >> "$log_file"
      
      # Create AMI
      AMI_ID=$(aws ec2 create-image \
        --region $REGION \
        --instance-id $INSTANCE_ID \
        --name "dumie-ami-from-$INSTANCE_ID" \
        --description "AMI created before terminating instance $INSTANCE_ID" \
        --no-reboot \
        --query 'ImageId' \
        --output text)
      
      if [ $? -eq 0 ]; then
        echo "$(date): Created AMI $AMI_ID" >> "$log_file"
        
        # Wait for AMI to be available
        aws ec2 wait image-available --region $REGION --image-ids $AMI_ID
        
        # Get the snapshot ID from the AMI
        SNAPSHOT_ID=$(aws ec2 describe-images \
          --region $REGION \
          --image-ids $AMI_ID \
          --query 'Images[0].BlockDeviceMappings[0].Ebs.SnapshotId' \
          --output text)
        
        if [ $? -eq 0 ]; then
          echo "$(date): Created snapshot $SNAPSHOT_ID from AMI" >> "$log_file"
          
          # Tag the snapshot
          aws ec2 create-tags \
            --region $REGION \
            --resources $SNAPSHOT_ID \
            --tags \
              "Key=Name,Value=$PROFILE" \
              "Key=InstanceID,Value=$INSTANCE_ID" \
              "Key=ManagedBy,Value=Dumie"
          
          # Deregister the AMI since we have the snapshot
          aws ec2 deregister-image --region $REGION --image-id $AMI_ID
          echo "$(date): Deregistered AMI $AMI_ID" >> "$log_file"
        fi
      fi

      # Release the lock before terminating the instance
      echo "$(date): Releasing lock before terminating instance..." >> "$log_file"
      release_lock "$LOCK_ID"
      
      # Terminate the instance
      aws ec2 terminate-instances --region $REGION --instance-ids $INSTANCE_ID
      echo "$(date): Terminated instance $INSTANCE_ID" >> "$log_file"
      
      exit 0
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
After=network.target sshd.service

[Service]
ExecStart=/usr/local/bin/ssh_monitor.sh
Restart=always
User=ec2-user

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable ssh-monitor.service
systemctl start ssh-monitor.service 
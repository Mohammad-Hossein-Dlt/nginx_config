#!/bin/bash
# Check for root privileges
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (sudo)."
    exit 1
fi

# Update package list
echo "Updating package list..."
apt-get update -y

# Install nginx
echo "Installing nginx..."
apt-get install nginx -y

# Create configuration file for port 80
CONFIG_FILE_80="/etc/nginx/conf.d/load_balancer_80.conf"
echo "Creating configuration file for port 80: $CONFIG_FILE_80"
cat > "$CONFIG_FILE_80" << 'EOF'

upstream backend_servers_80 {
    server 195.177.255.230:8000;
}

server {
    listen 80 default_server;
    server_name 195.177.255.230;

    location / {
        proxy_pass http://backend_servers_80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

# Create configuration file for port 8080
#CONFIG_FILE_8080="/etc/nginx/conf.d/load_balancer_8080.conf"
#echo "Creating configuration file for port 8080: $CONFIG_FILE_8080"
#cat > "$CONFIG_FILE_8080" << 'EOF'
#
#upstream backend_servers_8080 {
#    server 192.168.2.101;
#    server 192.168.2.102;
#}
#
#server {
#    listen 8080;
#    server_name yourdomain.com;  # Enter your domain name or appropriate IP
#
#    location / {
#        proxy_pass http://backend_servers_8080;
#        proxy_set_header Host $host;
#        proxy_set_header X-Real-IP $remote_addr;
#        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
#        proxy_set_header X-Forwarded-Proto $scheme;
#    }
#}
#EOF

# Test nginx configuration
echo "Testing nginx configuration..."
sudo nginx -t
if [ $? -ne 0 ]; then
    echo "Error in nginx configuration. Please check the config files."
    exit 1
fi

# Reload nginx to apply changes
echo "Reloading nginx..."
systemctl reload nginx

# Enable nginx service to automatically start on boot
echo "Enabling nginx service to automatically start after reboot..."
systemctl enable nginx

echo "Load balancer installation and configuration for ports 80 and 8080 completed successfully."

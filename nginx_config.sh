#!/bin/bash
# Check for root privileges
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (sudo)."
    exit 1
fi

# If nginx is installed, restore its default configuration
if [ -x "$(command -v nginx)" ]; then
    echo "Nginx is installed. Restoring default configuration..."
    # Remove additional configuration files in conf.d
    rm -f /etc/nginx/conf.d/*.conf

    # If using the sites-enabled structure, remove non-default sites
    if [ -d /etc/nginx/sites-enabled ]; then
        for site in /etc/nginx/sites-enabled/*; do
            if [ "$(basename "$site")" != "default" ]; then
                rm -f "$site"
            fi
        done
    fi

    # (Optional) Restore the original nginx.conf from backup if it exists
    if [ -f /etc/nginx/nginx.conf.backup ]; then
        cp /etc/nginx/nginx.conf.backup /etc/nginx/nginx.conf
    fi

    # Reload nginx to apply the changes
    systemctl reload nginx
fi

# Update package list
echo "Updating package list..."
apt-get update -y

# Install nginx
echo "Installing nginx..."
apt-get install nginx -y

# Create configuration file for port 80
CONFIG_FILE="/etc/nginx/conf.d/load_balancer.conf"
echo "Creating configuration file for port 80: $CONFIG_FILE"
cat > "$CONFIG_FILE" << 'EOF'

upstream load_balancer {
    server 195.177.255.230:8000;
}

server {
    listen 8080;
    server_name 195.177.255.230;

    location / {
        proxy_pass http://load_balancer;
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

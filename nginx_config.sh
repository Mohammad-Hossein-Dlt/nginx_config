#!/bin/bash
# Check for root privileges
if [ "$EUID" -ne 0 ]; then
    echo -e "\e[31m Please run as root (sudo). \e[0m"
    exit 1
fi

# If nginx is installed, restore its default configuration
if [ -x "$(command -v nginx)" ]; then
    echo -e "\e[32m Nginx is installed. Restoring default configuration... \e[0m"
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
echo -e "\e[32m Updating package list... \e[0m"
apt-get update -y

# Install nginx
echo -e "\e[32m Installing nginx... \e[0m"
apt-get install nginx -y

# Create configuration file for port 80
CONFIG_FILE="/etc/nginx/conf.d/load_balancer.conf"
echo -e "\e[32m Creating configuration file for port 80: $CONFIG_FILE \e[0m"
cat > "$CONFIG_FILE" << 'EOF'

upstream load_balancer {
    server 195.177.255.230:8000;
}

server {
    listen 80;
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

echo -e "\e[32m Removing default settings \e[0m"

sudo rm /etc/nginx/sites-enabled/default
sudo systemctl reload nginx

# Test nginx configuration
echo -e "\e[32m Testing nginx configuration... \e[0m"
sudo nginx -t
if [ $? -ne 0 ]; then
    echo -e "\e[32m Error in nginx configuration. Please check the config files. \e[0m"
    exit 1
fi

# Reload nginx to apply changes
echo -e "\e[32m Reloading nginx..."
sudo systemctl reload nginx

# Enable nginx service to automatically start on boot
echo -e "\e[32m Enabling nginx service to automatically start after reboot... \e[0m"
sudo systemctl enable nginx

echo -e "\e[32m Load balancer installation and configuration for ports 80 and 8080 completed successfully. \e[0m"

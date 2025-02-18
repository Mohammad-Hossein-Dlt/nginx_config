#!/bin/bash

colored_text(){
  local color=$1
  local text=$2
  echo -e "\e[${color}m$text\e[0m"
}

# Check for root privileges
if [ "$EUID" -ne 0 ]; then
    colored_text "31" "Please run as root (sudo)."
    exit 1
fi

# If nginx is installed, restore its default configuration
if [ -x "$(command -v nginx)" ]; then
    colored_text "32" "Nginx is installed.Purging existing installation and configuration files..."

    colored_text "32" "Stop nginx service..."
    sudo systemctl stop nginx 2>/dev/null
    colored_text "32" "Purging..."
    sudo apt-get purge -y nginx
    colored_text "32" "Purging..."
    sudo apt-get autoremove -y
    colored_text "32" "Removing nginx directory..."
    sudo rm -rf /etc/nginx

fi

# Update package list
colored_text "32" "Updating package list..."
apt-get update -y

# Install nginx
colored_text "32" "Installing nginx..."
apt-get install nginx -y

# Create configuration file for port 80
DOMAIN="hyperrio.site"
CONFIG_FILE="/etc/nginx/conf.d/load_balancer.conf"
colored_text "32" "\e[32m Creating configuration file for port 80: $CONFIG_FILE"
cat > "$CONFIG_FILE" << 'EOF'

upstream load_balancer {
    server 195.177.255.230:8000;
}

server {
    listen 80;
#    server_name 195.177.255.230;
    server_name hyperrio.site;

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

# Remove the default site configuration to avoid the welcome page
if [ -f /etc/nginx/sites-enabled/default ]; then
    colored_text "32" "Removing default settings at /etc/nginx/sites-enabled/default"
    rm -f /etc/nginx/sites-enabled/default
fi

# Remove any default configuration file in conf.d (e.g., default.conf)
if [ -f /etc/nginx/conf.d/default.conf ]; then
    colored_text "32" "Removing default settings at /etc/nginx/conf.d/default.conf"
    rm -f /etc/nginx/conf.d/default.conf
fi

# Test nginx configuration
colored_text "32" "Testing nginx configuration..."
sudo nginx -t
if [ $? -ne 0 ]; then
    colored_text "32" "Error in nginx configuration. Please check the config files."
    exit 1
fi

# Reload nginx to apply changes
colored_text "32" "Reloading nginx..."
sudo systemctl reload nginx

# Enable nginx service to automatically start on boot
colored_text "32" "Enabling nginx service to automatically start after reboot..."
sudo systemctl enable nginx

colored_text "32" "Load balancer installation and configuration completed successfully."

# Install ufw if not already installed
colored_text "32" "Installing firewall..."
sudo apt-get install -y ufw

# Allow SSH (port 22) to ensure remote access is not blocked
colored_text "32" "Allowing SSH on ports 80, 443"
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable ufw if it's not enabled already (this may prompt for confirmation)
sudo ufw --force enable

colored_text "32" "All is done."

#!/bin/bash

colored_text(){
  local color=$1
  local text=$2
  echo -e "\e[${color}m$text\e[0m"
}

function select_menu {
    options=("$@")

    select opt in "${options[@]}"; do
        case $opt in
            *)
                echo "$opt"
                break
                ;;
        esac
    done
}


# Check if the script is run as root
if [ "$EUID" -ne 0 ]; then
    colored_text "31" "Please run as root (sudo)."
    exit 1
fi

colored_text "36" "Fix the dpkg lock"
dpkg --configure -a

########################################
# Delete previous installations if any exist.
########################################

# If nginx is installed, remove its current installation and configuration files
if [ -x "$(command -v nginx)" ]; then
    colored_text "32" "Nginx is installed. Purging existing installation and configuration files..."

    colored_text "32" "Stopping nginx service..."
    systemctl stop nginx 2>/dev/null
    colored_text "32" "Purging nginx..."
    apt-get purge -y nginx
    colored_text "32" "Auto removing packages..."
    apt-get autoremove -y
    colored_text "32" "Removing nginx directory..."
    rm -rf /etc/nginx
fi

# If firewall is installed, remove its current installation and configuration files
if [ -x "$(command -v ufw)" ]; then
    colored_text "32" "Firewall is installed. Purging existing installation and configuration files..."
    ufw disable
    apt-get purge -y ufw
    apt-get autoremove -y ufw
    rm -rf /etc/ufw
fi

colored_text "32" "Removing previous ssl certificate..."
rm -rf /etc/ssl/certs/public_cert.crt
rm -rf /etc/ssl/private/private_key.key

########################################
# Update the package list and Install the required items
########################################

# Update the package list
colored_text "32" "Updating package list..."
apt-get update -y

# Install nginx
colored_text "32" "Installing nginx..."
apt-get install nginx -y
apt-get install -y ufw

########################################
# Get certificate and private key from user using nano
########################################

# Create a temporary file for certificate input
TMP_CERT=$(mktemp)
colored_text "36" "Please enter your certificate content in nano. Save and exit when done."
nano "$TMP_CERT"
CERTIFICATE_CONTENT=$(cat "$TMP_CERT")
rm -f "$TMP_CERT"

# Create a temporary file for private key input
TMP_KEY=$(mktemp)
colored_text "36" "Please enter your private key content in nano. Save and exit when done."
nano "$TMP_KEY"
PRIVATE_KEY_CONTENT=$(cat "$TMP_KEY")
rm -f "$TMP_KEY"

########################################
# Get Main Inputs From User
########################################

certification=$(select_menu "SSL" "No SSL")
setup=$(select_menu "Default" "Websocket")

certification=${certification: -No SSL}
setup=${setup: -Default}

colored_text "36" "$certification"
colored_text "36" "$setup"

########################################
# Domain and SSL Certificate Settings
########################################

DOMAIN="hyperrio.site"

# Paths where the certificate files will be saved
#CERT_PATH="/etc/ssl/certs/public_cert.crt"
#KEY_PATH="/etc/ssl/private/private_key.key"
BASE_PATH="/etc/ssl/files"

mkdir -p "$BASE_PATH"

CERT_PATH="$BASE_PATH/public_cert.crt"
KEY_PATH="$BASE_PATH/private_key.key"

# Create necessary directories if they do not exist
mkdir -p /etc/ssl/certs
mkdir -p /etc/ssl/private

# Save the certificate, private key, and chain file contents to their respective paths
echo "$CERTIFICATE_CONTENT" > "$CERT_PATH"
echo "$PRIVATE_KEY_CONTENT" > "$KEY_PATH"

########################################
# Nginx Configuration for Load Balancer and Reverse Proxy with SSL
########################################

CONFIG_FILE="/etc/nginx/conf.d/server.conf"
colored_text "32" "Creating configuration file for load balancer and reverse proxy: $CONFIG_FILE"

if [[ "$certification" = "SSL" && "$setup" = "Default" ]]; then
cat > "$CONFIG_FILE" <<EOF
# Define an upstream block for the backend server(s)
upstream load_balancer {
    server 195.177.255.230:8000;
}

# HTTP block: Redirect all HTTP traffic to HTTPS
server {
    listen 80;
    server_name ${DOMAIN};
    return 301 https://\$host\$request_uri;
}

# HTTPS block: SSL configuration and reverse proxy settings
server {
    listen 443 ssl;
    server_name ${DOMAIN};

    ssl_certificate ${CERT_PATH};
    ssl_certificate_key ${KEY_PATH};

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://load_balancer;

        proxy_set_header Host \$host;

        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}

EOF
elif [[ "$certification" = "SSL" && "$setup" = "Websocket" ]]; then
cat > "$CONFIG_FILE" <<EOF
# Define an upstream block for the backend server(s)
upstream load_balancer {
    ip_hash;
    server 195.177.255.230:8000;
}

# HTTP block: Redirect all HTTP traffic to HTTPS
server {
    listen 80;
    server_name ${DOMAIN};
    return 301 https://\$host\$request_uri;
}

# HTTPS block: SSL configuration and reverse proxy settings
server {
    listen 443 ssl;
    server_name ${DOMAIN};

    ssl_certificate ${CERT_PATH};
    ssl_certificate_key ${KEY_PATH};

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://load_balancer;

        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";

        proxy_set_header Host \$host;

        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF
elif [[ "$certification" = "No SSL" && "$setup" = "Default" ]]; then
cat > "$CONFIG_FILE" <<EOF
upstream load_balancer {
    server 195.177.255.230:8000;
}

server {
    listen 80;
    server_name 130.185.75.195;

    location / {
        proxy_pass http://load_balancer;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF
elif [[ "$certification" = "No SSL" && "$setup" = "Websocket" ]]; then
cat > "$CONFIG_FILE" <<EOF
upstream load_balancer {
    server 195.177.255.230:8000;
}

server {
    listen 80;
    server_name 130.185.75.195;

    location / {
        proxy_pass http://load_balancer;

        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";

        proxy_set_header Host \$host;

        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF
fi

# Remove default site configuration files if they exist
if [ -f /etc/nginx/sites-enabled/default ]; then
    colored_text "32" "Removing default configuration at /etc/nginx/sites-enabled/default"
    rm -f /etc/nginx/sites-enabled/default
fi

if [ -f /etc/nginx/conf.d/default.conf ]; then
    colored_text "32" "Removing default configuration at /etc/nginx/conf.d/default.conf"
    rm -f /etc/nginx/conf.d/default.conf
fi

# Test the nginx configuration
colored_text "32" "Testing nginx configuration..."
nginx -t
if [ $? -ne 0 ]; then
    colored_text "31" "Error in nginx configuration. Please check the config files."
    exit 1
fi

# Reload nginx to apply changes
colored_text "32" "Reloading nginx..."
systemctl reload nginx

# Enable nginx to start automatically on boot
colored_text "32" "Enabling nginx service to automatically start after reboot..."
systemctl enable nginx

colored_text "36" "Reverse proxy and Load balancer installation and configuration completed successfully."

########################################
# Firewall (ufw) Setup
########################################

colored_text "32" "Allowing SSH on port 22 and web traffic on ports 80, 443..."
ufw allow 9011/tcp
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp

# Enable ufw (this may prompt for confirmation)
ufw --force enable

colored_text "36" "All is done."

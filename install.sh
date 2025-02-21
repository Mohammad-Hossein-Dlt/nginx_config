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

########################################
# Get Main Inputs From User
########################################

certification=$(select_menu "SSL" "No SSL")
setup=$(select_menu "Default" "Websocket")

colored_text "36" "$certification"
colored_text "36" "$setup"

########################################
# Get certificate and private key from user using nano
########################################

DOMAIN="hyperrio.site"

BASE_PATH="/etc/ssl/files"
mkdir -p "$BASE_PATH"

function get_cert() {
    TMP_CERT=$(mktemp)
    colored_text "36" "Please enter your certificate content in nano. Save and exit when done."
    nano "$TMP_CERT" < /dev/tty > /dev/tty
    CERTIFICATE_CONTENT=$(cat "$TMP_CERT")
    rm -f "$TMP_CERT"

    CERT_PATH="$BASE_PATH/server.crt"
    echo "$CERTIFICATE_CONTENT" > "$CERT_PATH"
    echo "$CERT_PATH"
}

function get_key() {
    TMP_KEY=$(mktemp)
    colored_text "36" "Please enter your private key content in nano. Save and exit when done."
    nano "$TMP_KEY" < /dev/tty > /dev/tty
    PRIVATE_KEY_CONTENT=$(cat "$TMP_KEY")
    rm -f "$TMP_KEY"

    KEY_PATH="$BASE_PATH/server.key"
    echo "$PRIVATE_KEY_CONTENT" > "$KEY_PATH"
    echo "$KEY_PATH"
}

########################################
# Nginx Configuration for Load Balancer and Reverse Proxy
########################################

CONFIG_FILE="/etc/nginx/conf.d/server.conf"
colored_text "32" "Creating configuration file for load balancer and reverse proxy: $CONFIG_FILE"

if [[ "$certification" = "SSL" && "$setup" = "Default" ]]; then

CERT_PATH=$(get_cert)
KEY_PATH=$(get_key)

colored_text "32" "$CERT_PATH"
colored_text "32" "$KEY_PATH"

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

CERT_PATH=$(get_cert)
KEY_PATH=$(get_key)

colored_text "32" "$CERT_PATH"
colored_text "32" "$KEY_PATH"

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

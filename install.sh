#!/bin/bash

colored_text(){
  local color=$1
  local text=$2
  echo -e "\e[${color}m$text\e[0m" >&2
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

find_key_by_value() {
    local -n assoc_array=$1
    local search_value=$2

    for key in "${!assoc_array[@]}"; do
        if [ "${assoc_array[$key]}" == "$search_value" ]; then
            echo "$key"
            return 0  # موفقیت
        fi
    done

    echo ""
    return 1
}

########################################
# Initializing
########################################

# Check if the script is run as root
if [ "$EUID" -ne 0 ]; then
    colored_text "31" "Please run as root (sudo)."
    exit 1
fi

colored_text "36" "Fix the dpkg lock"
sudo kill 8001
dpkg --configure -a

colored_text "32" "Clear cache"
hash -r
rm -f management.shc
unset BASH_REMATCH
kill -9

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

CERT_BASE_PATH="/etc/ssl/files"
mkdir -p "$CERT_BASE_PATH"

DOMAIN="hyperrio.site"

function certificates() {

    # Directories to search for certificates
    directories=( "$CERT_BASE_PATH" )
    certificate_files=()

    # Find certificate files with common extensions (.crt, .pem, .cer)
    for dir in "${directories[@]}"; do
        if [ -d "$dir" ]; then
            while IFS= read -r file; do
                certificate_files+=("$file")
            done < <(find "$dir" -type f \( -iname "*.crt" -o -iname "*.pem" -o -iname "*.cer" \))
        fi
    done

    # Check if no certificates were found
    if [ ${#certificate_files[@]} -eq 0 ]; then
        echo "No certificates found."
        exit 1
    fi

    # Build menu options array with certificate details (excluding key file and path)
    declare -A items
    for cert in "${certificate_files[@]}"; do
        cert_file=$(basename "$cert")

        base_name="${cert_file%.*}"

        key_path="$CERT_BASE_PATH/${base_name}.key"
        key_file="$base_name.key"
        if [ ! -e "$key_path" ] && [ ! -f "$key_path" ]; then
            key_file="N/A"
        fi

        # Extract domain from the certificate subject (CN)
        domains=$(openssl x509 -in "$cert" -noout -ext subjectAltName 2>/dev/null | grep -o 'DNS:[^,]*' | sed 's/DNS://g' | paste -sd ", " -)

        if [ -z "$domains" ]; then
            domains="N/A"
        fi

        items["$base_name"]="Cert: $cert_file | Key: $key_file | Domains: $domains"
    done

    for key in "${!items[@]}"; do
        eval "$1[$key]=\"${items[$key]}\""
    done
}

function select_cert() {
    declare -A names
    certificates names
    selected=$(select_menu "${names[@]}")
    cert_path=$(find_key_by_value names "$selected")

    echo "$cert_path"
}

########################################
# Nginx Configuration for Load Balancer and Reverse Proxy
########################################

CONFIGS_BASE_PATH="/etc/nginx/conf.d"

colored_text "36" "Please enter a unique name for config file. previous configs show below:"
find "$CONFIGS_BASE_PATH" -type f -name "*.conf"
read -r name

colored_text "36" "Please enter the list of upstream IP addresses (space separated):"
read -r upstream_ips
IFS=' ' read -r -a upstream_array <<< "$upstream_ips"

# Build the upstream block based on the setup type (Websocket or Default)
upstream_conf="upstream load_balancer {"
if [[ "$setup" == "Websocket" ]]; then
    upstream_conf+="
    ip_hash;"
fi
for ip in "${upstream_array[@]}"; do
    upstream_conf+="
    server ${ip};"
done
upstream_conf+="
}"

CONFIG_FILE_PATH="$CONFIGS_BASE_PATH/${name}.conf"
colored_text "32" "Creating configuration file for load balancer and reverse proxy: $CONFIG_FILE_PATH"

if [[ "$certification" = "SSL" && "$setup" = "Default" ]]; then

selected_crt=$(select_cert)
CERT_PATH="$CERT_BASE_PATH/${selected_crt}.crt"
KEY_PATH="$CERT_BASE_PATH/${selected_crt}.key"

colored_text "32" "$CERT_PATH"
colored_text "32" "$KEY_PATH"

cat > "$CONFIG_FILE_PATH" <<EOF
# Define an upstream block for the backend server(s)
${upstream_conf}

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

selected_crt=$(select_cert)
CERT_PATH="$CERT_BASE_PATH/${selected_crt}.crt"
KEY_PATH="$CERT_BASE_PATH/${selected_crt}.key"

colored_text "32" "$CERT_PATH"
colored_text "32" "$KEY_PATH"

cat > "$CONFIG_FILE_PATH" <<EOF
# Define an upstream block for the backend server(s)
${upstream_conf}

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
cat > "$CONFIG_FILE_PATH" <<EOF
${upstream_conf}

server {
    listen 80;
    server_name 193.242.208.97;

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
cat > "$CONFIG_FILE_PATH" <<EOF
${upstream_conf}

server {
    listen 80;
    server_name 193.242.208.97;

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

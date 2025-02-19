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

colored_text "32" "Removing previous ssl certificate..."
rm -rf /etc/ssl/certs/public_cert.crt
rm -rf /etc/ssl/private/private_key.key

# Update the package list
colored_text "32" "Updating package list..."
apt-get update -y

# Install nginx
colored_text "32" "Installing nginx and fzf..."
apt-get install nginx -y

########################################
# Get Main Inputs From User
########################################

certification=$(select_menu "SSL" "No SSL")
setup=$(select_menu "Default" "Websocket")

colored_text "36" "$certification"
colored_text "36" "$setup"

########################################
# Domain and SSL Certificate Settings
########################################

DOMAIN="hyperrio.site"

# Directly assign the content of your certificate, private key, and chain file (if applicable)
CERTIFICATE_CONTENT="-----BEGIN CERTIFICATE-----
MIIEpjCCA46gAwIBAgIUeIRos63R4pCCM3bBI+0LVu9M2bEwDQYJKoZIhvcNAQEL
BQAwgYsxCzAJBgNVBAYTAlVTMRkwFwYDVQQKExBDbG91ZEZsYXJlLCBJbmMuMTQw
MgYDVQQLEytDbG91ZEZsYXJlIE9yaWdpbiBTU0wgQ2VydGlmaWNhdGUgQXV0aG9y
aXR5MRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRMwEQYDVQQIEwpDYWxpZm9ybmlh
MB4XDTI1MDIxOTA4MzAwMFoXDTI1MDIyNjA4MzAwMFowYjEZMBcGA1UEChMQQ2xv
dWRGbGFyZSwgSW5jLjEdMBsGA1UECxMUQ2xvdWRGbGFyZSBPcmlnaW4gQ0ExJjAk
BgNVBAMTHUNsb3VkRmxhcmUgT3JpZ2luIENlcnRpZmljYXRlMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu4O5g/QK0jWNSGuyCskhPXJHPHBFZJ3Eyous
4sr48GaDrGOxFx8GQmKW5LeNLOMu6IbK1sxvNFxCL/+fD068bJ4ZDrAo3TyPn1RK
l6zNJwfHmYihoIvIMu/cp8U9BPHddibYYPfKu1UuKIYDbTp5PWpVjyzOg+J12oS/
0uj8QTM5RmGRafBaHaExDB7eKp44xYnGvyQrrFRHb+kWhEL7sEH8hq/bGikGP9EN
acyO1Y0jWlcXRH0m6henA+nmer0gE2D4c9TmwVY/3suXhEX6taPkVn8u8tEBPcEL
GiRmYF6XwcjWLC6AlrkYGBYSeB+UA2P6HW0Cc7EKhTZnnAh/FwIDAQABo4IBKDCC
ASQwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcD
ATAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBTPqzSQZjYLncjqwwdwHBC6yysXwjAf
BgNVHSMEGDAWgBQk6FNXXXw0QIep65TbuuEWePwppDBABggrBgEFBQcBAQQ0MDIw
MAYIKwYBBQUHMAGGJGh0dHA6Ly9vY3NwLmNsb3VkZmxhcmUuY29tL29yaWdpbl9j
YTApBgNVHREEIjAggg8qLmh5cGVycmlvLnNpdGWCDWh5cGVycmlvLnNpdGUwOAYD
VR0fBDEwLzAtoCugKYYnaHR0cDovL2NybC5jbG91ZGZsYXJlLmNvbS9vcmlnaW5f
Y2EuY3JsMA0GCSqGSIb3DQEBCwUAA4IBAQCBZpSxex28s9PVaC3DCrf42Qg5aJyH
Ht0fncN2Dde9XVPI1RhCLvTEmq9JNhkD9hAKVaoWrju7wOKNC/Mtbf2F4OlKS1Z3
BwzIwnEe8gSgk00Ch+A9rdH0Dc6rylEGWSXp8trQl1njeh7PIHTGQInDyhrRP/C/
8eKD6mxsUOQvnm2JIKKhDMns9AetaxE1KAcYoVrUbqPX2ZfwRe62IilMyNgq54Ku
bC3EQ2UbLnHX3DAkvPWJiC4RTnEyutsfg8UrvWoXitdj6tZkNIBdnukjH2sAD7t9
L6zVM91HxGhajOby4EP4p7dag5djCZknZZ8zw/DIUdL9yB32o292A2gQ
-----END CERTIFICATE-----"

PRIVATE_KEY_CONTENT="-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7g7mD9ArSNY1I
a7IKySE9ckc8cEVkncTKi6ziyvjwZoOsY7EXHwZCYpbkt40s4y7ohsrWzG80XEIv
/58PTrxsnhkOsCjdPI+fVEqXrM0nB8eZiKGgi8gy79ynxT0E8d12Jthg98q7VS4o
hgNtOnk9alWPLM6D4nXahL/S6PxBMzlGYZFp8FodoTEMHt4qnjjFica/JCusVEdv
6RaEQvuwQfyGr9saKQY/0Q1pzI7VjSNaVxdEfSbqF6cD6eZ6vSATYPhz1ObBVj/e
y5eERfq1o+RWfy7y0QE9wQsaJGZgXpfByNYsLoCWuRgYFhJ4H5QDY/odbQJzsQqF
NmecCH8XAgMBAAECggEAAhNU+osfkr1GNHbOTRfcbVPjFyrEkEOKyeJX5iUjy9oz
boIMxybwKJkmXTiG47TnPQQMIYQe98mZor+SlCNlkVI13TEn+QlSXnIbAHWVY7w0
6McklHFXpIpKuyfcvNd5TkIWnQqN5oFTD6tdq4s9JYF6zVxMNBUanT7HT/jr0EQe
OMrGyRUYOhYdnnIdz6ENO8dBaD/TXjidbzF3dxi2mDsSJb8fR5pgEHCrFBN/moRz
HdujM1up7zfQC+WcBjDo+vbb4cQvnbFBhYWatI1lZkaW9lRVUDNXZa65Dq9rXZhP
OSD4L9DB59ZahLFQSCLUJND47hp0X7SrnzsHFRqvMQKBgQDo5bqQBPyVOopKfpIW
/zqAVHgnFxtTlKzrMRuU0Cb4npCE0Si8hv5goTTvmWn0ocNON3o9mqttlsvJw/fS
b4H9MHff1f1yOlifA927TYqebWHrswZKWSaT6MDdiFTx5ZM6ROpI4lOAddNH9+a2
cv4VqFklFx4TDTBqXimzgI+L7wKBgQDOHYeav2jlbpRJharld+DE7I40nhInnxsq
V1ReJLHlO5h9R2VNL2pJt3p6S7MbpMVfuIMew7MzMPESyFxC1uHuWOjMKukbWtoM
2mQBJLKgFgHen1hE5am42JfhIG3Sj2C1s5KiwCuP58azvEzxNluf78miLdbrQAUP
/S1Xz3e3WQKBgB7GXISzY/0EI0n8t6k6SKy0fLwNnZrJxp+9eXuMldm8ejRSvyNK
Y0q5gpk2mH2u8nPfeNOzIHv2tS8QKisweOjQAscdK8RwWU2J4T7i3DJbGdlfarFg
XmylEPc1EKR2RaIpgRvobEhJSYX1CBOL1m9eM4lnKJ4z2/XyQ2ho0I8ZAoGAYt+L
g1I6sYSgIby7RCSDcDPB67/AGb2bPG50DE0yATLbbY1oLOSH6iDX4f6aRrJ98/MB
AysBtZbOriHrEC0gaEPCON6EwBiO7Qd+XYYLIfwsnWx23WYGSqOsB9SUmiMpU0B3
IRdqTjfy+5lil3tp7IkMgn3W0Tb+trLOo4bkeNECgYEAgMaL4FFcO9l88J7UH0TX
et/t539L6Y53mOvyJLW5QW4Z+U2DBwsItBLkvmiX+MjOb2+FfbmL7FaywNuGuBrR
X5JxikKQs5ZogEH1GHs4Ctmu8ZhXt6myb2VrpkKQQa1wppw+HC7cAWLdBVeY+fdG
x5JuyV14hD/WA+acV5pB4M4=
-----END PRIVATE KEY-----"

# Paths where the certificate files will be saved
CERT_PATH="/etc/ssl/certs/public_cert.crt"
KEY_PATH="/etc/ssl/private/private_key.key"

# Create necessary directories if they do not exist
mkdir -p /etc/ssl/certs
mkdir -p /etc/ssl/private

# Save the certificate, private key, and chain file contents to their respective paths
echo "$CERTIFICATE_CONTENT" > "$CERT_PATH"
echo "$PRIVATE_KEY_CONTENT" > "$KEY_PATH"

########################################
# Nginx Configuration for Load Balancer and Reverse Proxy with SSL
########################################

CONFIG_FILE="/etc/nginx/conf.d/load_balancer.conf"
colored_text "32" "Creating configuration file for load balancer and reverse proxy: $CONFIG_FILE"

if [[ "$certification" = "SSL" && "$setup" = "Default" ]]; then
colored_text "31" "jehfierhfg49p8ghp9o34tjgiolerwtjvnipretuhgb9p35thgrtojgi9op3rithg9354tg"
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

colored_text "32" "Installing firewall..."
apt-get install -y ufw

colored_text "32" "Allowing SSH on port 22 and web traffic on ports 80, 443..."
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp

# Enable ufw (this may prompt for confirmation)
ufw --force enable

colored_text "36" "All is done."

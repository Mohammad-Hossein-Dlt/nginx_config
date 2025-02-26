#!/bin/bash

colored_text(){
  local color=$1
  local text=$2
  echo -e "\e[${color}m$text\e[0m" >&2
}

#function select_menu {
#    options=("$@")
#
#    select opt in "${options[@]}"; do
#        case $opt in
#            *)
#                echo "$opt"
#                break
#                ;;
#        esac
#    done
#}



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

function select_menu() {

    # توابع کمکی برای کنترل چاپ ترمینال و دریافت کلید
    ESC=$(printf "\033")
    cursor_blink_on()  { printf "%s\n" "${ESC}[?25h"; }
    cursor_blink_off() { printf "%s\n" "${ESC}[?25l"; }
    cursor_to()        { printf "%s\n" "${ESC}[$1;${2:-1}H"; }
    print_option()     { printf "%s\n" "   $1 "; }
    print_selected()   { printf "%s\n" "  ${ESC}[7m $1 ${ESC}[27m"; }
    get_cursor_row()   {
      IFS=';' read -r -sdR -p $'\E[6n' ROW COL; echo "${ROW#*[}";
      export COL
      }
    key_input()        {
        read -r -s -n3 key 2>/dev/null
        if [[ $key == "${ESC}[A" ]]; then echo up; fi
        if [[ $key == "${ESC}[B" ]]; then echo down; fi
        if [[ -z $key ]]; then echo enter; fi
    }

    # چاپ چند خط خالی اولیه (جهت اسکرول کردن در صورت پایین بودن صفحه)
    for opt; do printf "\n"; done

    # تعیین موقعیت فعلی کرسر برای نوشتن مجدد گزینه‌ها
    local last_row
    last_row=$(get_cursor_row)
    local start_row=$(( last_row - $# ))

    # اطمینان از روشن شدن کرسر و بازگرداندن حالت echo در صورت قطع برنامه با ctrl+c
    trap "cursor_blink_on; stty echo; printf '\n'; exit" SIGINT
    cursor_blink_off

    local selected=0
    while true; do
        # چاپ گزینه‌ها با بازنویسی خطوط آخر
        local idx=0
        for opt; do
            cursor_to $(( start_row + idx ))
            if [ $idx -eq $selected ]; then
                print_selected "$opt"
            else
                print_option "$opt"
            fi
            ((idx++))
        done

        # کنترل کلید‌های ورودی کاربر
        case $(key_input) in
            enter) break ;;
            up)    ((selected--))
                   if [ $selected -lt 0 ]; then selected=$(( $# - 1 )); fi ;;
            down)  ((selected++))
                   if [ $selected -ge $# ]; then selected=0; fi ;;
        esac
    done

    # بازگرداندن موقعیت کرسر به حالت عادی
    cursor_to "$last_row"
    printf "\n"
    cursor_blink_on

    return $selected
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
# Get certification & setup Inputs From User
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

#DOMAIN="hyperrio.site"

function certificates() {
    local -n arr_ref=$1

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
        colored_text "93" "No certificates found."
        exit 1
    fi

    # Build menu options array with certificate details
    for cert in "${certificate_files[@]}"; do
        cert_file=$(basename "$cert")

        base_name="${cert_file%.*}"

        key_path="$CERT_BASE_PATH/${base_name}.key"
        key_file="$base_name.key"
        if [ ! -e "$key_path" ] && [ ! -f "$key_path" ]; then
            key_file="N/A"
        fi

        # Extract domain from the certificate subject (CN)
        domains=$(openssl x509 -in "$cert" -noout -ext subjectAltName 2>/dev/null | grep -o 'DNS:[^,]*' | sed 's/DNS://g' | paste -sd " // " -)

        if [ -z "$domains" ]; then
            domains="N/A"
        fi

        arr_ref["$base_name"]="Cert: $cert_file | Key: $key_file | Domains: $domains"
    done

    export arr_ref
}

function select_cert() {
    declare -A names
    certificates names
    if [ $? -ne 0 ]; then
        exit 1
    fi
    selected=$(select_menu "${names[@]}")
    cert_path=$(find_key_by_value names "$selected")

    echo "$cert_path"
}

extract_dns() {
    local cert_file="$1"
    local ref="$2"
    declare -n dns_list="$ref"
    dns_list=()

    while IFS= read -r line; do
        if [[ -n "$line" ]]; then
            dns_list+=("$line")
        fi
    done < <(openssl x509 -in "$cert_file" -noout -ext subjectAltName 2>/dev/null | grep -o 'DNS:[^,]*' | sed 's/DNS://g')
}

########################################
# Nginx Configuration for Load Balancer and Reverse Proxy
########################################

CONFIGS_BASE_PATH="/etc/nginx/conf.d"

function get_domain() {
    colored_text "32" "Please enter your domain:"
    read -r domain
    echo "$domain"
}

function get_ip() {
    colored_text "32" "Please enter the ip of this server:"
    read -r domain
    echo "$domain"
}

function get_http_port() {
    colored_text "32" "Please enter http port (80 is default):"
    read -r port
    port=${port:-80}
    echo "$port"
}

function get_https_port() {
    colored_text "32" "Please enter https port (443 is default):"
    read -r port
    port=${port:-443}
    echo "$port"
}

colored_text "36" "Please enter a unique name for config file. previous configs show below:"
find "$CONFIGS_BASE_PATH" -type f -name "*.conf"
read -r name

if [[ -f "$CONFIGS_BASE_PATH/${name}.conf" ]]; then
    colored_text "93" "The name you entered already exist."
    exit 1
fi

colored_text "36" "Please enter the list of upstream IP addresses (space separated):"
read -r upstream_ips
IFS=' ' read -r -a upstream_array <<< "$upstream_ips"

upstream_conf="upstream ${name} {"
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
colored_text "32" "Creating configuration file for load balancer and reverse proxy:"

if [[ "$certification" = "SSL" && "$setup" = "Default" ]]; then

selected_crt=$(select_cert)
if [ $? -ne 0 ]; then
    exit 1
fi
declare -a domains
extract_dns "$CERT_BASE_PATH/${selected_crt}.crt" domains
selected_domain=$(select_menu "${domains[@]}")
CERT_PATH="$CERT_BASE_PATH/${selected_crt}.crt"
KEY_PATH="$CERT_BASE_PATH/${selected_crt}.key"

HTTP_PORT=$(get_http_port)
HTTPS_PORT=$(get_https_port)

cat > "$CONFIG_FILE_PATH" <<EOF
# Define an upstream block for the backend server(s)
${upstream_conf}

# HTTP block: Redirect all HTTP traffic to HTTPS
server {
    listen ${HTTP_PORT};
    server_name ${selected_domain};
    return 301 https://\$host\$request_uri;
}

# HTTPS block: SSL configuration and reverse proxy settings
server {
    listen ${HTTPS_PORT} ssl;
    server_name ${selected_domain};

    ssl_certificate ${CERT_PATH};
    ssl_certificate_key ${KEY_PATH};

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://${name};

        proxy_set_header Host \$host;

        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}

EOF
elif [[ "$certification" = "SSL" && "$setup" = "Websocket" ]]; then

selected_crt=$(select_cert)
if [ $? -ne 0 ]; then
    exit 1
fi
declare -a domains
extract_dns "$CERT_BASE_PATH/${selected_crt}.crt" domains
selected_domain=$(select_menu "${domains[@]}")
CERT_PATH="$CERT_BASE_PATH/${selected_crt}.crt"
KEY_PATH="$CERT_BASE_PATH/${selected_crt}.key"

HTTP_PORT=$(get_http_port)
HTTPS_PORT=$(get_https_port)

cat > "$CONFIG_FILE_PATH" <<EOF
# Define an upstream block for the backend server(s)
${upstream_conf}

# HTTP block: Redirect all HTTP traffic to HTTPS
server {
    listen ${HTTP_PORT};
    server_name ${selected_domain};
    return 301 https://\$host\$request_uri;
}

# HTTPS block: SSL configuration and reverse proxy settings
server {
    listen ${HTTPS_PORT} ssl;
    server_name ${selected_domain};

    ssl_certificate ${CERT_PATH};
    ssl_certificate_key ${KEY_PATH};

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://${name};

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

server_ip=$(get_ip)
HTTP_PORT=$(get_http_port)

cat > "$CONFIG_FILE_PATH" <<EOF
${upstream_conf}

server {
    listen ${HTTP_PORT};
    server_name ${server_ip};

    location / {
        proxy_pass http://${name};
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF
elif [[ "$certification" = "No SSL" && "$setup" = "Websocket" ]]; then

server_ip=$(get_ip)
HTTP_PORT=$(get_http_port)

cat > "$CONFIG_FILE_PATH" <<EOF
${upstream_conf}

server {
    listen ${HTTP_PORT};
    server_name ${server_ip};

    location / {
        proxy_pass http://${name};

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

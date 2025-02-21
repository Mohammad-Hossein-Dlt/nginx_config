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


# Check if the script is run as root
if [ "$EUID" -ne 0 ]; then
    colored_text "31" "Please run as root (sudo)."
    exit 1
fi

colored_text "36" "Fix the dpkg lock"
sudo kill 8001
dpkg --configure -a

########################################
# Nginx Management
########################################

function install_nginx() {
    colored_text "32" "Updating package list..."
    apt-get update -y
    colored_text "32" "Installing nginx..."
    apt-get install nginx -y
}

function delete_nginx() {
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
}

function configs() {
    config_files=$(find /etc/nginx/conf.d/ -type f -name "*.conf" -exec basename {} \;)
    echo "${config_files[@]}"
}

function delete_config() {
    file_name=$1
    if [[ -n "$file_name" ]]; then
        rm -rf /etc/nginx/conf.d/"$file_name"
        colored_text "32" "Config ${file_name} deleted."
    else
        colored_text "31" "Can not delete directory conf.d in path /etc/nginx/conf.d"
    fi
}


function edit_config() {
    file_name=$1
    nano /etc/nginx/conf.d/"$file_name"
    colored_text "32" "Config ${file_name} edited."
}

########################################
# Firewall Management
########################################

function firewall_status() {
    sudo ufw status
}

function install_firewall() {
    colored_text "32" "Installing firewall..."
    apt-get install -y ufw
}
function delete_firewall() {
    if [ -x "$(command -v ufw)" ]; then
        colored_text "32" "Firewall is installed. Purging existing installation and configuration files..."
        ufw disable
        apt-get purge -y ufw
        apt-get autoremove -y ufw
        rm -rf /etc/ufw
    fi
}

function opening_ports() {
    read -p "Enter ports to open (separated by space): " -a ports
    for port in "${ports[@]}"; do
        colored_text "32" "Opening port: $port"
        sudo ufw allow "$port"/tcp
    done
    colored_text "32" " Check firewall status"
    sudo ufw status
}

########################################
# Certificate Management
########################################

function certificates() {
    # Directories to search for certificates
#    directories=( "/etc/ssl/certs" "/etc/ssl/private" "/etc/pki/tls/certs" "/etc/pki/tls/private" "/etc/letsencrypt/live" )
    directories=( "/etc/ssl/files" )
    certificate_files=()

    # Find certificate files with common extensions (.crt, .pem, .cer)
    for dir in "${directories[@]}"; do
        if [ -d "$dir" ]; then
            while IFS= read -r file; do
                certificate_files+=("$file")
            done < <(find "$dir" -type f \( -iname "*.crt" -o -iname "*.key" -o -iname "*.pem" -o -iname "*.cer" \))
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

        # Extract domain from the certificate subject (CN)
        domains=$(openssl x509 -in "$cert" -noout -ext subjectAltName 2>/dev/null | grep -o 'DNS:[^,]*' | sed 's/DNS://g' | paste -sd ", " -)

        if [ -z "$domains" ]; then
            domains="N/A"
        fi

        items["$cert"]="Path: $cert_file | Domains: $domains"
    done

    for key in "${!items[@]}"; do
        eval "$1[$key]=\"${items[$key]}\""
    done
}

function certificate_info() {
    local cert_path=$1

    cert_file=$(basename "$cert_path")
    domains=$(openssl x509 -in "$cert_path" -noout -ext subjectAltName 2>/dev/null | grep -o 'DNS:[^,]*' | sed 's/DNS://g' | paste -sd ", " -)

    # Extract certificate validity dates using openssl
    dates=$(openssl x509 -in "$cert_path" -noout -dates 2>/dev/null)
    notBefore=$(echo "$dates" | grep 'notBefore=' | cut -d'=' -f2)
    notAfter=$(echo "$dates" | grep 'notAfter=' | cut -d'=' -f2)

    # Fallback if dates cannot be extracted
    if [ -z "$notBefore" ]; then
        notBefore="N/A"
    fi
    if [ -z "$notAfter" ]; then
        notAfter="N/A"
    fi

    # Build the menu option string with certificate file name and validity dates
    info="Cert: $cert_file \n Domains: $domains \n Valid from: $notBefore \n Valid to: $notAfter"
    colored_text "36" "$info"
}

function delete_certificate() {
    colored_text "32" "Removing previous ssl certificate..."
    rm -rf /etc/ssl/certs/public_cert.crt
    rm -rf /etc/ssl/private/private_key.key
}

########################################
# Install requirements
########################################

function install_requirements() {
    colored_text "32" "Updating package list..."
    apt-get update -y

    colored_text "32" "Installing nginx..."
    apt-get install nginx -y
    colored_text "32" "Installing firewall..."
    apt-get install -y ufw
}

########################################
# Menu
########################################

colored_text "32" "Management menu"
opt=$(select_menu "Install Requirements" "Nginx Management" "Firewall Management" "Certificate Management")

if [ "$opt" = "Install Requirements" ]; then
    
    install_requirements

elif [ "$opt" = "Nginx Management" ]; then
    nginx_opt=$(select_menu "Install Nginx" "Delete Nginx" "Add Config" "Manage Configs")

    if [ "$nginx_opt" = "Install Nginx" ]; then
        install_nginx
    elif [ "$nginx_opt" = "Delete Nginx" ];then
        delete_nginx
    elif [ "$nginx_opt" = "Add Config" ];then
        bash <(curl -Ls https://raw.githubusercontent.com/Mohammad-Hossein-Dlt/nginx_config/master/install_v2.sh)
    elif [ "$nginx_opt" = "Manage Configs" ];then
        colored_text "36" "test"
        files=$(configs)
        selected_file=$(select_menu "${files[@]}")
        config_opt=$(select_menu "Delete Config" "Edit Config")
        if [ "$config_opt" = "Delete Config" ]; then
            delete_config "$selected_file"
        elif [ "$config_opt" = "Edit Config" ]; then
            edit_config "$selected_file"
        fi
    fi

elif [ "$opt" = "Firewall Management" ]; then
    firewall_opt=$(select_menu "Firewall Status" "Install Firewall" "Delete Firewall" "Open port(s)")

    if [ "$firewall_opt" = "Firewall Status" ]; then
        firewall_status
    elif [ "$firewall_opt" = "Install Firewall" ]; then
        install_firewall
    elif [ "$firewall_opt" = "Delete Firewall" ];then
        delete_firewall
    elif [ "$firewall_opt" = "Open port(s)" ];then
        opening_ports
    fi

elif [ "$opt" = "Certificate Management" ]; then

    declare -A names
    certificates names

    selected=$(select_menu "${names[@]}")

    cert_path=$(find_key_by_value names "$selected")

    colored_text "36" "$cert_path"

    certificate_opt=$(select_menu "Certificate Info" "Delete Certificate")

    if [ "$certificate_opt" = "Certificate Info" ]; then
        certificate_info "$cert_path"
    fi

fi



if [ -x "$(command -v nginx)" ]; then
    systemctl reload nginx
fi


hash -r
rm -f management.shc
unset BASH_REMATCH
kill -9
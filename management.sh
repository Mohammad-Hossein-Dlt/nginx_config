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

function select_certificate() {
    # Directories to search for certificates
    directories=( "/etc/ssl/certs" "/etc/pki/tls/certs" "/etc/letsencrypt/live" )
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
    menu_options=()
    for cert in "${certificate_files[@]}"; do
        cert_file=$(basename "$cert")

        # Extract domain from the certificate subject (CN)
        domain=$(openssl x509 -in "$cert" -noout -ext subjectAltName 2>/dev/null | grep -o 'DNS:[^,]*' | sed 's/DNS://g' | paste -sd ", " -)

        if [ -z "$domain" ]; then
            domain="N/A"
        fi

        menu_options+=("Cert: $cert_file | Domains: $domain" )
    done

#    for opt in "${menu_options[@]}"; do
#        echo "$opt"
#    done
#    for cert in "${certificate_files[@]}"; do
#        cert_file=$(basename "$cert")
#
#        # Extract certificate validity dates using openssl
#        dates=$(openssl x509 -in "$cert" -noout -dates 2>/dev/null)
#        notBefore=$(echo "$dates" | grep 'notBefore=' | cut -d'=' -f2)
#        notAfter=$(echo "$dates" | grep 'notAfter=' | cut -d'=' -f2)
#
#        # Fallback if dates cannot be extracted
#        if [ -z "$notBefore" ]; then
#            notBefore="N/A"
#        fi
#        if [ -z "$notAfter" ]; then
#            notAfter="N/A"
#        fi
#
#        # Build the menu option string with certificate file name and validity dates
#        menu_options+=("Cert: $cert_file | Valid from: $notBefore | Valid to: $notAfter")
#    done
    echo "${menu_options[@]}"

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

#    mapfile -t names < <(select_certificate)

    names=$(select_certificate)
    selected=$(select_menu "${names[@]}")

    colored_text "36" "$selected"

fi



if [ -x "$(command -v nginx)" ]; then
    systemctl reload nginx
fi
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
# Nginx Management
########################################

function install_nginx() {
    colored_text "32" "Updating package list..."
    apt-get update -y
    colored_text "32" "Installing nginx..."
    apt-get install nginx -y
}

function uninstall_nginx() {
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
    colored_text "32" "Updating package list..."
    apt-get update -y
    colored_text "32" "Installing firewall..."
    apt-get install -y ufw

    if [ -x "$(command -v ufw)" ]; then
        ufw allow 9011/tcp
        ufw allow 22/tcp
    fi
}
function uninstall_firewall() {
    if [ -x "$(command -v ufw)" ]; then
        colored_text "32" "Firewall is installed. Purging existing installation and configuration files..."
        colored_text "32" "Stopping ufw service..."
        ufw disable
        colored_text "32" "Purging firewall ufw..."
        apt-get purge -y ufw
        colored_text "32" "Auto removing packages..."
        apt-get autoremove -y ufw
        colored_text "32" "Removing ufw directory..."
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

        local base_name="${cert_file%.*}"

        # Extract domain from the certificate subject (CN)
        domains=$(openssl x509 -in "$cert" -noout -ext subjectAltName 2>/dev/null | grep -o 'DNS:[^,]*' | sed 's/DNS://g' | paste -sd ", " -)

        if [ -z "$domains" ]; then
            domains="N/A"
        fi

        key_path="/etc/ssl/files/${base_name}.key"
        key_file="$base_name.key"
        if [ ! -e "$key_path" ] && [ ! -f "$key_path" ]; then
            key_file="N/A"
        fi
        items["$cert"]="Cert: $cert_file | Key: $key_file | Domains: $domains"
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
    info="Cert: $cert_file\nKey: $key_file\nDomains: $domains\nValid from: $notBefore\nValid to: $notAfter"
    colored_text "36" "$info"
}

function delete_certificate() {
    local file_path=$1
    cert_file=$(basename "$file_path")

    local base_name="${cert_file%.*}"

    colored_text "32" "Removing ssl certificate '$base_name'"
    rm -rf "/etc/ssl/files/${base_name}.crt"
    rm -rf "/etc/ssl/files/${base_name}.key"
}

function delete_all_certificate() {
    colored_text "32" "Removing all ssl certificates..."
    rm -rf "/etc/ssl/files/*.crt"
    rm -rf "/etc/ssl/files/*.key"
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
# Uninstall requirements
########################################

function uninstall_everything() {
    uninstall_nginx
    uninstall_firewall
    delete_all_certificate
}

########################################
# Menu
########################################

colored_text "32" "Management menu"
opt=$(select_menu "Install Requirements" "Nginx Management" "Firewall Management" "Certificate Management" "Reinstall everything" "Uninstall and delete everything")

if [ "$opt" = "Install Requirements" ]; then
    
    install_requirements

elif [ "$opt" = "Nginx Management" ]; then
    nginx_opt=$(select_menu "Install Nginx" "Delete Nginx" "Add Config" "Manage Configs")

    if [ "$nginx_opt" = "Install Nginx" ]; then
        install_nginx
    elif [ "$nginx_opt" = "Delete Nginx" ];then
        colored_text -n "94" "Do you really want to uninstall nginx? yes or y to confirm no or n to cancel."
        read confirm
        if [[ "${confirm,,}" = "yes" || "${confirm,,}" = "y" ]];then
            uninstall_nginx
        else
            colored_text "93" "Uninstall nginx canceled."
        fi
    elif [ "$nginx_opt" = "Add Config" ];then
        if [ ! -x "$(command -v nginx)" ]; then
            colored_text "31" "Nginx not installed."
        elif [ ! -x "$(command -v ufw)" ]; then
            colored_text "31" "Firewall not installed."
        else
            bash <(curl -Ls https://raw.githubusercontent.com/Mohammad-Hossein-Dlt/nginx_config/master/install.sh)
        fi
    elif [ "$nginx_opt" = "Manage Configs" ];then
        colored_text "36" "test"
        files=$(configs)
        selected_file=$(select_menu "${files[@]}")
        config_opt=$(select_menu "Delete Config" "Edit Config")
        if [ "$config_opt" = "Delete Config" ]; then
            colored_text -n "94" "Do you really want to delete config $selected_file? yes or y to confirm no or n to cancel."
            read confirm
            if [[ "${confirm,,}" = "yes" || "${confirm,,}" = "y" ]];then
                delete_config "$selected_file"
            else
                colored_text "93" "Delete config $selected_file canceled."
            fi
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
        colored_text -n "94" "Do you really want to uninstall firewall (ufw)? yes or y to confirm no or n to cancel."
        read confirm
        if [[ "${confirm,,}" = "yes" || "${confirm,,}" = "y" ]];then
            uninstall_firewall
        else
            colored_text "93" "Uninstall firewall (ufw) canceled."
        fi
    elif [ "$firewall_opt" = "Open port(s)" ];then
        opening_ports
    fi

elif [ "$opt" = "Certificate Management" ]; then

    declare -A names
    certificates names

    selected=$(select_menu "${names[@]}")

    cert_path=$(find_key_by_value names "$selected")

    certificate_opt=$(select_menu "Certificate Info" "Delete Certificate")

    if [ "$certificate_opt" = "Certificate Info" ]; then
        certificate_info "$cert_path"
    elif [ "$certificate_opt" = "Delete Certificate" ]; then
        colored_text -n "94" "Do you really want to delete certificate $cert_path? yes or y to confirm no or n to cancel."
        read confirm
        if [[ "${confirm,,}" = "yes" || "${confirm,,}" = "y" ]];then
            delete_certificate "$cert_path"
        else
            colored_text "93" "Delete certificate $cert_path canceled."
        fi
    fi

elif [ "$opt" = "Reinstall everything" ]; then

    colored_text -n "94" "Do you really want to reinstall everything? yes or y to confirm no or n to cancel."
    read confirm
    if [[ "${confirm,,}" = "yes" || "${confirm,,}" = "y" ]];then
        uninstall_nginx
        install_nginx
        uninstall_firewall
        install_firewall
        delete_all_certificate
        install_requirements
        bash <(curl -Ls https://raw.githubusercontent.com/Mohammad-Hossein-Dlt/nginx_config/master/install.sh)
    else
        colored_text "93" "Reinstall everything canceled."
    fi

elif [ "$opt" = "Uninstall and delete everything" ]; then
    colored_text -n "94" "Do you really want to uninstall and delete everything? yes or y to confirm no or n to cancel."
    read confirm
    if [[ "${confirm,,}" = "yes" || "${confirm,,}" = "y" ]];then
        uninstall_everything
    else
        colored_text "93" "Uninstall and delete everything canceled."
    fi

fi

if [ -x "$(command -v nginx)" ]; then
    systemctl reload nginx
fi

colored_text "32" "Clear cache"
hash -r
rm -f management.shc
unset BASH_REMATCH
kill -9
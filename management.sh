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
    config_files=$(find /etc/nginx/conf.d -type f -name "*.conf" -exec basename {} \;)
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

CERT_BASE_PATH="/etc/ssl/files"
mkdir -p "$CERT_BASE_PATH"

function certificates() {
    # Directories to search for certificates
#    directories=( "/etc/ssl/certs" "/etc/ssl/private" "/etc/pki/tls/certs" "/etc/pki/tls/private" "/etc/letsencrypt/live" )
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

        items["$cert"]="Cert: $cert_file | Key: $key_file | Domains: $domains"
    done

    for key in "${!items[@]}"; do
        eval "$1[$key]=\"${items[$key]}\""
    done
}

function certificate_info() {
    local cert_path=$1

    cert_file=$(basename "$cert_path")

    base_name="${cert_file%.*}"

    key_path="$CERT_BASE_PATH/${base_name}.key"
    key_file="$base_name.key"

    if [ ! -e "$key_path" ] && [ ! -f "$key_path" ]; then
        key_file="N/A"
    fi

    domains=$(openssl x509 -in "$cert_path" -noout -ext subjectAltName 2>/dev/null | grep -o 'DNS:[^,]*' | sed 's/DNS://g' | paste -sd ", " -)

    if [ -z "$domains" ]; then
        domains="N/A"
    fi

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


function get_cert() {
    local name=$1
    TMP_CERT=$(mktemp)
    colored_text "36" "Please enter your certificate content in nano. Save and exit when done." >&2
    nano "$TMP_CERT" < /dev/tty > /dev/tty
    CERTIFICATE_CONTENT=$(cat "$TMP_CERT")
    rm -f "$TMP_CERT"

    CERT_PATH="$CERT_BASE_PATH/$name.crt"
    echo "$CERTIFICATE_CONTENT" > "$CERT_PATH"
    echo "$CERT_PATH"
}

function get_key() {
    local name=$1
    TMP_KEY=$(mktemp)
    colored_text "36" "Please enter your private key content in nano. Save and exit when done." >&2
    nano "$TMP_KEY" < /dev/tty > /dev/tty
    PRIVATE_KEY_CONTENT=$(cat "$TMP_KEY")
    rm -f "$TMP_KEY"

    KEY_PATH="$CERT_BASE_PATH/$name.key"
    echo "$PRIVATE_KEY_CONTENT" > "$KEY_PATH"
    echo "$KEY_PATH"
}

function delete_certificate() {
    local file_path=$1
    cert_file=$(basename "$file_path")

    local base_name="${cert_file%.*}"

    colored_text "32" "Removing ssl certificate '$base_name'"
    rm -rf "$CERT_BASE_PATH/${base_name}.crt"
    rm -rf "$CERT_BASE_PATH/${base_name}.key"
}

function delete_all_certificate() {
    colored_text "32" "Removing all ssl certificates..."
    rm -rf $CERT_BASE_PATH/*.crt
    rm -rf $CERT_BASE_PATH/*.key
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
    opt=$(select_menu "Install Nginx" "Delete Nginx" "Add Config" "Manage Configs")

    if [ "$opt" = "Install Nginx" ]; then
        install_nginx
    elif [ "$opt" = "Delete Nginx" ];then
        colored_text "94" "Do you really want to uninstall nginx? yes or y to confirm no or n to cancel."
        read confirm
        if [[ "${confirm,,}" = "yes" || "${confirm,,}" = "y" ]];then
            uninstall_nginx
        else
            colored_text "93" "Uninstall nginx canceled."
        fi
    elif [ "$opt" = "Add Config" ];then
        if [ ! -x "$(command -v nginx)" ]; then
            colored_text "31" "Nginx not installed."
        elif [ ! -x "$(command -v ufw)" ]; then
            colored_text "31" "Firewall not installed."
        else
            bash <(curl -Ls https://raw.githubusercontent.com/Mohammad-Hossein-Dlt/nginx_config/master/install.sh)
        fi
    elif [ "$opt" = "Manage Configs" ];then
        files=$(configs)
        selected_file=$(select_menu "${files[@]}")
        opt=$(select_menu "Delete Config" "Edit Config")
        if [ "$opt" = "Delete Config" ]; then
            colored_text "94" "Do you really want to delete config $selected_file? yes or y to confirm no or n to cancel."
            read confirm
            if [[ "${confirm,,}" = "yes" || "${confirm,,}" = "y" ]];then
                delete_config "$selected_file"
            else
                colored_text "93" "Delete config $selected_file canceled."
            fi
        elif [ "$opt" = "Edit Config" ]; then
            edit_config "$selected_file"
        fi
    fi

elif [ "$opt" = "Firewall Management" ]; then
    opt=$(select_menu "Firewall Status" "Install Firewall" "Delete Firewall" "Open port(s)")

    if [ "$opt" = "Firewall Status" ]; then
        firewall_status
    elif [ "$opt" = "Install Firewall" ]; then
        install_firewall
    elif [ "$opt" = "Delete Firewall" ];then
        colored_text "94" "Do you really want to uninstall firewall (ufw)? yes or y to confirm no or n to cancel."
        read confirm
        if [[ "${confirm,,}" = "yes" || "${confirm,,}" = "y" ]];then
            uninstall_firewall
        else
            colored_text "93" "Uninstall firewall (ufw) canceled."
        fi
    elif [ "$opt" = "Open port(s)" ];then
        opening_ports
    fi

elif [ "$opt" = "Certificate Management" ]; then

    opt=$(select_menu "Add Certificate" "Delete All Certificates" "Manage a Certificate")

    if [ "$opt" = "Add Certificate" ]; then
        colored_text "94" "Enter certificate file name. the name must be unique"
        read name
        CERT_PATH=$(get_cert "$name")
        KEY_PATH=$(get_key "$name")

        if [[ -n "$CERT_PATH" && -n "$KEY_PATH" ]]; then
            colored_text "32" "Certificate $name created."
        fi
    elif [ "$opt" = "Delete All Certificates" ]; then
        colored_text "94" "Do you really want to delete all certificates? yes or y to confirm no or n to cancel."
        read confirm
        if [[ "${confirm,,}" = "yes" || "${confirm,,}" = "y" ]];then
            delete_all_certificate
        else
            colored_text "93" "Delete all certificates canceled."
        fi
    elif [ "$opt" = "Manage a Certificate" ]; then
        declare -A names
        certificates names

        selected=$(select_menu "${names[@]}")

        cert_path=$(find_key_by_value names "$selected")

        certificate_opt=$(select_menu "Certificate Info" "Delete Certificate")

        if [ "$certificate_opt" = "Certificate Info" ]; then
            certificate_info "$cert_path"
        elif [ "$certificate_opt" = "Delete Certificate" ]; then
            colored_text "94" "Do you really want to delete certificate $cert_path? yes or y to confirm no or n to cancel."
            read confirm
            if [[ "${confirm,,}" = "yes" || "${confirm,,}" = "y" ]];then
                delete_certificate "$cert_path"
            else
                colored_text "93" "Delete certificate $cert_path canceled."
            fi
        fi

    fi

elif [ "$opt" = "Reinstall everything" ]; then

    colored_text "94" "Do you really want to reinstall everything? yes or y to confirm no or n to cancel."
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
    colored_text "94" "Do you really want to uninstall and delete everything? yes or y to confirm no or n to cancel."
    read confirm
    if [[ "${confirm,,}" = "yes" || "${confirm,,}" = "y" ]];then
        uninstall_everything
    else
        colored_text "93" "Uninstall and delete everything canceled."
    fi

fi

if [ -x "$(command -v nginx)" ]; then
    systemctl reload nginx
    nginx -t
fi

colored_text "32" "Clear cache"
hash -r
rm -f management.shc
unset BASH_REMATCH
kill -9
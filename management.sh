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
    nginx_opt=$(select_menu "Install Nginx" "Delete Nginx" "Manage Configs")

    if [ "$nginx_opt" = "Install Nginx" ]; then
        install_nginx
    elif [ "$nginx_opt" = "Delete Nginx" ];then
        delete_nginx
    elif [ "$nginx_opt" = "Manage Configs" ];then
        colored_text "36" "test"
        files=$(configs)
        select_menu $files
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
    pass
fi
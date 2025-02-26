#!/bin/bash

colored_text(){
  local color=$1
  local text=$2
  echo -e "\e[${color}m$text\e[0m" >&2
}

# Check if Go is installed
if ! command -v go &> /dev/null
then
    colored_text "32" "Go is not installed, installing Go..."
    sudo apt update
    sudo apt install -y golang-go

else
    colored_text "32" "Go is already installed. Removing Go..."
    sudo apt remove -y golang-go
    sudo apt-get purge -y nginx

    colored_text "32" "Installing Go..."
    sudo apt update
    sudo apt install -y golang-go
fi

if ! command -v go &> /dev/null
then
  colored_text "31" "Error installing Go!"
  exit 1
fi

colored_text "32" "Initializing go mod..."
curl -Ls https://raw.githubusercontent.com/Mohammad-Hossein-Dlt/nginx_config/master/go.mod | tee go.mod | go mod tidy
colored_text "32" "Run..."
curl -Ls https://raw.githubusercontent.com/Mohammad-Hossein-Dlt/nginx_config/master/main.go | go run /dev/stdin
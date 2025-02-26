#!/bin/bash

colored_text(){
  local color=$1
  local text=$2
  echo -e "\e[${color}m$text\e[0m" >&2
}

TARGET_DIR="/home/avida_cli"
REPO_URL="Mohammad-Hossein-Dlt/nginx_config"

# Check if Go is installed
if ! command -v go &> /dev/null
then
    colored_text "32" "Go is not installed, installing Go..."
    sudo apt update
    sudo apt install -y golang-go
else
    colored_text "32" "Go is already installed. Removing Go..."
    sudo apt remove -y golang-go
    sudo apt-get purge -y golang-go

    colored_text "32" "Installing Go..."
    sudo apt update
    sudo apt install -y golang-go
fi

if ! command -v go &> /dev/null
then
  colored_text "31" "Error installing Go!"
  exit 1
fi

# Check if Git is installed
if ! command -v git &> /dev/null
then
    colored_text "32" "Git is not installed, installing Git..."
    sudo apt update
    sudo apt install -y git
else
    colored_text "32" "Git is already installed. Removing Git..."
    sudo apt remove -y git
    sudo apt-get purge -y git

    colored_text "32" "Installing Git..."
    sudo apt update
    sudo apt install -y git
fi

if ! command -v git &> /dev/null
then
  colored_text "31" "Error installing Git!"
  exit 1
fi

#
#colored_text "32" "Initializing go mod..."
#curl -Ls https://raw.githubusercontent.com/Mohammad-Hossein-Dlt/nginx_config/master/go.mod | tee go.mod | go mod tidy
#colored_text "32" "Run..."
#curl -Ls https://raw.githubusercontent.com/Mohammad-Hossein-Dlt/nginx_config/master/main.go | go run /dev/stdin

git clone https://github.com/$REPO_URL "$TARGET_DIR"

cd "$TARGET_DIR" || exit

go mod tidy

go run main.go


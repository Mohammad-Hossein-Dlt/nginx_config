
### Install
```Bash
bash <(curl -Ls https://raw.githubusercontent.com/Mohammad-Hossein-Dlt/nginx_config/master/install.sh)
```

### Management
```Bash
bash <(curl -Ls https://raw.githubusercontent.com/Mohammad-Hossein-Dlt/nginx_config/master/management.sh)
```

### Management
```Bash
tmpfile=$(mktemp) && trap "rm -f $tmpfile" EXIT && curl -L -o "$tmpfile" https://raw.githubusercontent.com/Mohammad-Hossein-Dlt/nginx_config/master/nginx_configure && chmod +x "$tmpfile" && "$tmpfile"
```


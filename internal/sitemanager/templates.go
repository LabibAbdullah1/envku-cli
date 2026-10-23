package sitemanager

import (
	"fmt"
	"strings"
)

// GenerateNginxConfig renders the appropriate Nginx server block based on site configuration
func GenerateNginxConfig(cfg SiteConfig) string {
	switch cfg.Type {
	case SiteTypeNode:
		return generateNodeTemplate(cfg)
	case SiteTypeStatic:
		return generateStaticTemplate(cfg)
	case SiteTypePHP:
		fallthrough
	default:
		return generatePHPTemplate(cfg)
	}
}

// generatePHPTemplate generates an optimized Nginx server block for PHP-FPM applications
func generatePHPTemplate(cfg SiteConfig) string {
	phpVer := cfg.PHPVersion
	if phpVer == "" {
		phpVer = "8.3"
	}

	docRoot := strings.TrimRight(cfg.DocumentRoot, "/")
	if docRoot == "" {
		docRoot = fmt.Sprintf("/var/www/%s/public", cfg.Domain)
	}

	return fmt.Sprintf(`# Configured by Envku CLI
server {
    listen 80;
    listen [::]:80;
    server_name %s;
    root %s;

    index index.php index.html index.htm;
    charset utf-8;

    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # Max Upload Size
    client_max_body_size 64M;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location = /favicon.ico { access_log off; log_not_found off; }
    location = /robots.txt  { access_log off; log_not_found off; }

    error_page 404 /index.php;

    location ~ \.php$ {
        fastcgi_pass unix:/run/php/php%s-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        include fastcgi_params;
        fastcgi_hide_header X-Powered-By;
        fastcgi_read_timeout 300;
    }

    location ~ /\.(?!well-known).* {
        deny all;
    }

    # Gzip Compression
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
}
`, cfg.Domain, docRoot, phpVer)
}

// generateNodeTemplate generates a reverse proxy Nginx server block for Node.js / SSR apps
func generateNodeTemplate(cfg SiteConfig) string {
	port := cfg.NodeProxyPort
	if port <= 0 {
		port = 3000
	}

	return fmt.Sprintf(`# Configured by Envku CLI
server {
    listen 80;
    listen [::]:80;
    server_name %s;

    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;

    client_max_body_size 64M;

    location / {
        proxy_pass http://127.0.0.1:%d;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 300;
        proxy_connect_timeout 300;
    }

    location ~ /\.(?!well-known).* {
        deny all;
    }
}
`, cfg.Domain, port)
}

// generateStaticTemplate generates a server block for static sites (HTML / React / Vue SPA)
func generateStaticTemplate(cfg SiteConfig) string {
	docRoot := strings.TrimRight(cfg.DocumentRoot, "/")
	if docRoot == "" {
		docRoot = fmt.Sprintf("/var/www/%s", cfg.Domain)
	}

	return fmt.Sprintf(`# Configured by Envku CLI
server {
    listen 80;
    listen [::]:80;
    server_name %s;
    root %s;

    index index.html index.htm;
    charset utf-8;

    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location ~* \.(jpg|jpeg|png|gif|ico|css|js|woff2|woff|ttf|svg)$ {
        expires 30d;
        add_header Cache-Control "public, no-transform";
    }

    location ~ /\.(?!well-known).* {
        deny all;
    }

    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml text/javascript;
}
`, cfg.Domain, docRoot)
}

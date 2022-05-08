# Adequate Auth
A self-hosted user authentication server for use with NGINX's `auth_request` module. Designed as a simple replacement for `basic_auth`, which can be cumbersome and less secure.

# Getting Started

First, ensure your server has a valid ssl certificate and NGINX is configured to use it. I recommend using [certbot](https://certbot.eff.org/) to manage certificates on your server.

## Installation
Note: Your file paths may vary depending on operating system and configuration.

* Download the latest release or compile from source  
* Copy `better_auth` and `static/login.html` to `/opt/better_auth/`
* Copy `nginx/better_auth` to `/etc/nginx/sites-enabled/`
* Add `include sites-enabled/better_auth` to NGINX server entries that should be protected, eg:

```
server {
        listen 443 default_server;
        listen [::]:443 default_server;

        ssl_certificate /etc/letsencrypt/live/my.site.url/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/my.site.url/privkey.pem;
        include /etc/letsencrypt/options-ssl-nginx.conf;
        ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

        server_name _;
        location /secret_hideout{
                proxy_pass http://localhost:1234/;
        }
        
        include sites-enabled/better_auth;
}
```

The addition of the last `include` line is all that is required for this server config.

## Adding users
Before `better_auth` will run, a user and password must be added. Run the following command [using your own username and password] to add users:
```
/opt/better_auth/better_auth adduser MegaMan87 dR.7#0m4$.7i8#t
```
Users can be removed by deleting their corresponding line in your `better_auth.users` file.
If a user is added while `better_auth` is running, `better_auth` will need to be restarted to recognize the user.

## Config
Running `better_auth` for the first time will generate a config file `/etc/better_auth/better_auth.conf`, which is a simple JSON-style config. The default settings will be sufficient for most users, but may be changed to anything you prefer.

* `ServerAddress`: ip address on which the server will listen [`localhost`]
* `ServerPort`: port number on which the server will listen [`8675`]
* `SessionTimeout`: time in seconds after which an inactive session will expire, requiring the user to log in again [`3600`]
* `AuthFile`: file containing users and passwords entered via `adduser` [`/etc/better_auth/better_auth.conf`]
* `LogFile`: file containing log information [`/var/log/better_auth.log`]

<b>Note:</b> Changing `ServerAddress` or `ServerPort` will require corresponding changes to be made to `/etc/nginx/sites-enabled/adequte_auth` so NGINX knows where to send requests.

## Starting better_auth automatically

### Systemd
* Copy service file
```
cp -v /opt/better_auth/runscripts/better_auth.service /etc/systemd/system/better_auth.service
```
* Set permissions
```
sudo chown root:root /etc/systemd/system/better_auth.service
sudo chmod 644 /etc/systemd/system/better_auth.service
```
* Enable and start
```
sudo systemctl enable better_auth
sudo systemctl start better_auth
```


# How it Works
In any nginx `server` block containing `better_auth`, nginx will ask `better_auth` if the current user is logged in. If not, the user is presented with the login page. If the user enters a valid username and password `better_auth` starts a new session for the user. A random session-token is generated and sent to the user as a cookie and the user is sent to the originally-requested page. Any time a user requests a new page the cookie containing their session-token is sent to `better_auth`. If the session-token is valid and has not expired nginx is allowed to continue with the request. Otherwise, the user is again presented with the login page to sign in.

Usernames and passwords are stored in the users file on individual lines as `username:hashed_password`. This is a similar format to a typical `.htpasswd` file, but `better_auth` passwords are hashed using `bcrypt` and cannot be reasonably un-hashed by any force currently known to man.


# koito-multi-proxy

Proxies requests based on the token provided in the authorization header.

95% AI code. Use at your own risk.

## Usage
docker compose
```yml
services:
  koito-multi-proxy:
    image: gabehf/koito-multi-proxy:latest
    container_name: koito-multi-proxy
    ports:
      - "4111:4111"
    volumes:
      - /my/config/dir:/etc/kmp
```
Then put your config.yml into `/my/config/dir/config.yml`

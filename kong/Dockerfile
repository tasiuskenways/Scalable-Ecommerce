FROM kong:3.9.1

USER root

# Copy plugin into Kong’s Lua path
COPY plugins/ /usr/local/share/lua/5.1/kong/plugins/

USER kong

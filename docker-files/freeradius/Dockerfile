FROM freeradius/freeradius-server:latest

COPY mods-available/sql /etc/freeradius/mods-available/sql
COPY mods-available/rest /etc/freeradius/mods-available/rest
COPY mods-available/expiration /etc/freeradius/mods-available/expiration

COPY sites-enabled/default /etc/freeradius/sites-enabled/default

COPY dictionary /etc/freeradius/dictionary

# Symlink mods-enabled
RUN ln -sf ../mods-available/sql /etc/freeradius/mods-enabled/sql && \
    ln -sf ../mods-available/rest /etc/freeradius/mods-enabled/rest && \
    ln -sf ../mods-available/expiration /etc/freeradius/mods-enabled/expiration

CMD ["radiusd", "-X"]

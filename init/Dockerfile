FROM alpine:3.21

# Install dependencies: sqlite3, wget, unzip, bash
RUN apk add --no-cache sqlite sqlite-dev wget unzip bash moreutils

WORKDIR /app

# Copy SQL and entrypoint
COPY init.sql /app/init.sql
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

VOLUME /data

ENTRYPOINT ["/app/entrypoint.sh"]

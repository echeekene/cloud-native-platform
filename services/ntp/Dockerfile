FROM alpine:3.20
RUN apk add --no-cache chrony tzdata \
 && mkdir -p /var/lib/chrony /var/log/chrony /run/chrony \
 && chown -R chrony:chrony /var/lib/chrony /var/log/chrony /run/chrony
COPY chrony.conf /etc/chrony/chrony.conf
COPY healthcheck.sh /usr/local/bin/chrony-healthcheck
EXPOSE 123/udp
# -d stay in foreground, -x never touch the host clock (serve only),
# -u drop privileges to the chrony user after binding port 123
ENTRYPOINT ["chronyd", "-d", "-x", "-u", "chrony", "-f", "/etc/chrony/chrony.conf"]

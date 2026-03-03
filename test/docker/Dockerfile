# Pi-Star MCP test container
# Installs all MMDVM APK packages + the dashboard binary on Alpine 3.22
FROM alpine:3.22

# Add Pi-Star APK repository
RUN wget -qO /etc/apk/keys/hamradio.rsa.pub https://apk.pistar.uk/hamradio.rsa.pub \
 && echo "https://apk.pistar.uk/v3.22/community" >> /etc/apk/repositories

# Install all MMDVM packages + mosquitto
RUN apk update && apk add --no-cache \
    mmdvmhost \
    dmrclients \
    ysfclients \
    p25clients \
    nxdnclients \
    dstarclients \
    dstarrepeater \
    fmclients \
    aprsclients \
    pocsagclients \
    mosquitto

# Create Pi-Star dashboard directories
RUN mkdir -p \
    /etc/pistar-dashboard/certs \
    /var/lib/pistar-dashboard/backups \
    /var/log/pistar-dashboard \
    /run/pistar

# Copy dashboard binary
COPY pistar-dashboard-linux-arm64 /usr/local/bin/pistar-dashboard
RUN chmod +x /usr/local/bin/pistar-dashboard

# Copy dev config
COPY dashboard.ini /etc/pistar-dashboard/dashboard.ini

# Copy sample MMDVM.ini so the configurator has something to read/write
COPY MMDVM.ini /etc/mmdvmhost/MMDVM.ini
COPY DMRGateway.ini /etc/dmrclients/DMRGateway.ini
COPY YSFGateway.ini /etc/ysfclients/YSFGateway.ini

# Activate .example configs so radio config fan-out has real files to write
RUN for f in \
      /etc/ysfclients/YSF2DMR.ini \
      /etc/ysfclients/YSF2NXDN.ini \
      /etc/ysfclients/YSF2P25.ini \
      /etc/ysfclients/DGIdGateway.ini \
      /etc/dmrclients/DMR2YSF.ini \
      /etc/dmrclients/DMR2NXDN.ini \
      /etc/nxdnclients/NXDN2DMR.ini \
      /etc/nxdnclients/NXDNGateway.ini \
      /etc/p25clients/P25Gateway.ini \
      /etc/fmclients/FMGateway.ini \
      /etc/aprsclients/APRSGateway.ini \
      /etc/pocsagclients/DAPNETGateway.ini \
    ; do \
      [ -f "$f.example" ] && cp "$f.example" "$f" ; \
    done

# Expose HTTP and HTTPS
EXPOSE 8080 8443

# Run the dashboard
CMD ["/usr/local/bin/pistar-dashboard", "--config", "/etc/pistar-dashboard/dashboard.ini"]

# 使用 debian:trixie-slim（Debian 13），与主 Dockerfile runtime 阶段的 base 镜像对齐。
#
# 历史说明：之前固定 bookworm 是因为 trixie 对 armhf 做了 64-bit time_t ABI 迁移，
# 把 libcups2/libcupsimage2/libpng16-16 等包重命名为 t64 后缀版本。现在的做法是：
# ① 不再显式安装这些低级库包——它们都是 cups/cups-filters 的依赖，apt 会自动
#    根据架构拉取正确的包名（如 libcups2t64、libpng16-16t64 等）；
# ② printer-driver-gutenprint 在 trixie armhf 上已无 binary（gutenprint 源码包
#    未对 armhf 构建），因此改为条件安装——仅在 amd64/arm64 上装入。
FROM debian:trixie-slim

ARG TARGETARCH
ARG TARGETVARIANT

LABEL maintainer="涵曦 <im.hanxi@gmail.com>"
LABEL description="Debian-based CUPS server with Epson L380 official driver preinstalled"

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Asia/Shanghai
ENV CUPSADMIN=print
ENV CUPSPASSWORD=print

# 安装 CUPS 核心及常用打印驱动。
# 低级库包（libcups2、libcupsimage2、libpng16-16、libjpeg62-turbo、libtiff6、
# libusb-1.0-0）都是 cups/cups-filters 的传递依赖，apt 会根据架构自动解析到
# 正确的包名（trixie armhf 上带 t64 后缀），因此不再显式列出。
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      wget \
      curl \
      gnupg2 \
      ca-certificates \
      cups \
      cups-client \
      cups-filters \
      lsb-release \
      apt-utils \
      usbutils \
      printer-driver-all \
      printer-driver-cups-pdf \
      printer-driver-escpr \
      printer-driver-foo2zjs \
      foomatic-db-compressed-ppds \
      openprinting-ppds \
      hpijs-ppds \
      hp-ppd \
      hplip \
      avahi-daemon && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# printer-driver-gutenprint 在 trixie armhf 上没有 binary 包（gutenprint 的
# 链接依赖 libgutenprint9 未完成 t64 迁移），仅在 amd64/arm64 上安装。
# armhf 用户仍可通过 printer-driver-all 推荐的其他驱动覆盖大部分打印机。
RUN if [ "${TARGETARCH}" != "arm" ]; then \
      apt-get update && \
      apt-get install -y --no-install-recommends printer-driver-gutenprint && \
      apt-get clean && rm -rf /var/lib/apt/lists/*; \
    fi

# Only install Epson proprietary driver on amd64; other architectures will use the Debian drivers
RUN if [ "${TARGETARCH}" = "amd64" ]; then \
    mkdir -p /tmp/epson_driver && \
    wget -q -O /tmp/epson_driver/epson-inkjet-printer-201601w_1.0.1-1_amd64.deb \
    https://download-center.epson.com.cn/f/module/81976474-cb4d-4b75-a0d8-ebc9ebde70d0/epson-inkjet-printer-201601w_1.0.1-1_amd64.deb && \
    wget -q -O /tmp/epson_driver/epson-printer-utility_1.2.2-1_amd64.deb \
    https://download-center.epson.com.cn/f/module/b04dc5b0-cb78-4ce7-98c8-06731fb74062/epson-printer-utility_1.2.2-1_amd64.deb && \
    dpkg -i /tmp/epson_driver/*.deb || apt-get install -y -f --no-install-recommends; \
    rm -rf /tmp/epson_driver; \
  fi && rm -rf /var/lib/apt/lists/*

RUN sed -i 's/Listen localhost:631/Listen 0.0.0.0:631/' /etc/cups/cupsd.conf && \
    sed -i 's/Browsing Off/Browsing On/' /etc/cups/cupsd.conf && \
    sed -i 's/<Location \/>/<Location \/>\n  Allow All/' /etc/cups/cupsd.conf && \
    sed -i 's/<Location \/admin>/<Location \/admin>\n  Allow All\n  Require user @SYSTEM/' /etc/cups/cupsd.conf && \
    sed -i 's/<Location \/admin\/conf>/<Location \/admin\/conf>\n  Allow All/' /etc/cups/cupsd.conf && \
    echo "ServerAlias *" >> /etc/cups/cupsd.conf && \
    echo "DefaultEncryption Never" >> /etc/cups/cupsd.conf
RUN cp -rp /etc/cups /etc/cups-bak
VOLUME [ "/etc/cups" ]

EXPOSE 631

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

CMD ["/entrypoint.sh"]

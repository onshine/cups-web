FROM oven/bun AS frontend-build
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN bun install
COPY frontend ./
RUN bun run build

# ---- Java OFD converter build ----
FROM maven:3.9-eclipse-temurin-17 AS java-builder
WORKDIR /src/ofd-converter
COPY ofd-converter/pom.xml ./
RUN mvn dependency:go-offline -q
COPY ofd-converter/src ./src
RUN mvn clean package -q -DskipTests

FROM golang:1.26 AS builder
WORKDIR /src

# copy go modules and source
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY . .
# Copy built frontend assets into expected location for go:embed
COPY --from=frontend-build /src/frontend/dist ./frontend/dist

# Build the Go binary (frontend must be built before this step in CI/local)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -o /out/cups-web ./cmd/server

FROM debian:bookworm-slim AS runtime

# Install LibreOffice (headless conversion), Ghostscript, and minimal fonts/certificates
#
# 关于 fonts-droid-fallback（重要）：
#   Ghostscript 的 pdfwrite 在处理"空壳 CJK 字体"（Type0 + UniGB-UCS2-H 但无 FontFile、
#   无 ToUnicode、原字体名为 GBK 字节如 "黑体"/"宋体"）时，会按 Adobe-GB1 CID 查找替身字体。
#   gs 内建的兜底查找路径是 /usr/share/ghostscript/<ver>/Resource/CIDFSubst/DroidSansFallback.ttf
#   —— 这是一个按 Adobe-GB1 CID 编号的 CJK 字体，恰好与 UniGB-UCS2-H 推导的 CID 对齐。
#   Debian 出于软件包分拆把这个字体从 ghostscript 主包中剥离到独立的 fonts-droid-fallback 包；
#   ghostscript 主包里的 CIDFSubst/DroidSansFallback.ttf 是指向 fonts-droid-fallback 的软链接，
#   不装这个包就会得到"空壳字体被替换到错误字体"导致的中文乱码
#   （macOS brew 的 ghostscript 主包自带这个字体，所以本地测试不会踩到；Docker 里才会）。
#   诊断方法：gs -dPDFDEBUG <in.pdf> 2>&1 | grep "Loading CIDFont"
#   若看到 "substitute from .../DroidSansFallback.ttf" 说明字体替换路径正常。
#
# fonts-noto-cjk / fonts-arphic-* / fonts-wqy-zenhei 负责 LibreOffice headless 渲染 Office 文档
# 时的中文字形；它们按 Unicode 组织，不能替代 DroidSansFallback 在 gs CIDFSubst 路径的角色。
RUN apt-get update && apt-get install -y --no-install-recommends \
    libreoffice-core libreoffice-writer libreoffice-calc libreoffice-impress openjdk-17-jre \
    ghostscript fonts-droid-fallback \
    fonts-dejavu-core fonts-noto-cjk fonts-arphic-uming fonts-arphic-ukai fonts-wqy-zenhei \
    ca-certificates \
  && rm -rf /var/lib/apt/lists/*

# Create a non-root user for running the service
RUN groupadd -r nonroot && useradd -r -g nonroot nonroot

RUN mkdir -p \
    /home/nonroot/.cache/dconf \
    /home/nonroot/.config/libreoffice \
    /home/nonroot/.local/share/libreoffice \
  && chown -R nonroot:nonroot /home/nonroot/ \
  && chmod -R 755 /home/nonroot/ \
  && chmod 700 /home/nonroot/.cache/dconf

ENV DCONF_USER_CONFIG_DIR=/home/nonroot/.config/dconf
ENV HOME=/home/nonroot
ENV XDG_CACHE_HOME=/home/nonroot/.cache

COPY --from=builder /out/cups-web /cups-web
COPY --from=java-builder /src/ofd-converter/target/ofd-converter.jar /ofd-converter.jar
EXPOSE 8080
USER nonroot
ENTRYPOINT ["/cups-web"]

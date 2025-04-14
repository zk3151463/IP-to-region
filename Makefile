APP_NAME := ip-location-service
MAIN_FILE := main.go
GO := go
PORT := 8080
GEOIP_DB := GeoLite2-City.mmdb
IP2REGION_DB := ip2region.xdb
LICENSE_KEY := #密钥

.PHONY: all build clean run help download-geoip update-geoip build-linux download-ip2region update-ip2region

all: build

help:
	@echo "Make commands:"
	@echo "  make build          - 构建应用"
	@echo "  make build-linux    - 构建Linux平台应用"
	@echo "  make clean         - 清理构建文件"
	@echo "  make run           - 运行应用 (默认端口: $(PORT))"
	@echo "  make run PORT=xxx  - 在指定端口运行应用"
	@echo "  make run DB=xxx    - 使用指定数据库类型运行应用 (geoip 或 ip2region)"
	@echo "  make download-geoip - 下载 GeoIP 数据库"
	@echo "  make update-geoip   - 更新 GeoIP 数据库"
	@echo "  make download-ip2region - 下载 ip2region 数据库"
	@echo "  make update-ip2region   - 更新 ip2region 数据库"

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin/data/ip2region
	$(GO) build -o bin/$(APP_NAME) $(MAIN_FILE)
	@if [ -f data/$(GEOIP_DB) ]; then \
		cp data/$(GEOIP_DB) bin/data/; \
		echo "Copied GeoIP database to bin/data/"; \
	fi
	@if [ -f data/ip2region/$(IP2REGION_DB) ]; then \
		cp data/ip2region/$(IP2REGION_DB) bin/data/ip2region/; \
		echo "Copied ip2region database to bin/data/ip2region/"; \
	fi
	@echo "Build complete!"

build-linux:
	@echo "Building $(APP_NAME) for Linux..."
	@mkdir -p bin/linux/data/ip2region
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o bin/linux/$(APP_NAME) $(MAIN_FILE)
	@if [ -f data/$(GEOIP_DB) ]; then \
		cp data/$(GEOIP_DB) bin/linux/data/; \
		echo "Copied GeoIP database to bin/linux/data/"; \
	fi
	@if [ -f data/ip2region/$(IP2REGION_DB) ]; then \
		cp data/ip2region/$(IP2REGION_DB) bin/linux/data/ip2region/; \
		echo "Copied ip2region database to bin/linux/data/ip2region/"; \
	fi
	@echo "Linux build complete!"

clean:
	@echo "Cleaning..."
	@rm -rf bin
	@go clean
	@echo "Clean complete!"

run: build
	@echo "Running $(APP_NAME) on port $(PORT)..."
	@if [ "$(DB)" = "ip2region" ]; then \
		./bin/$(APP_NAME) -port=$(PORT) -db-type=ip2region -ip2region-db=bin/data/ip2region/$(IP2REGION_DB); \
	else \
		./bin/$(APP_NAME) -port=$(PORT) -db-type=geoip -geoip-db=bin/data/$(GEOIP_DB); \
	fi

# 下载 GeoIP 数据库
download-geoip:
	@echo "Downloading GeoIP database..."
	@echo "Using license key: $(LICENSE_KEY)"
	@echo "Downloading database file..."
	@mkdir -p data
	@if ! curl -L -v -o data/geoip.tar.gz "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=$(LICENSE_KEY)&suffix=tar.gz"; then \
		echo "Error: Failed to download GeoIP database. Please check your license key and network connection."; \
		exit 1; \
	fi
	@echo "Extracting database file..."
	@if ! tar -xzf data/geoip.tar.gz -C data; then \
		echo "Error: Failed to extract database file."; \
		rm -f data/geoip.tar.gz; \
		exit 1; \
	fi
	@echo "Moving database file..."
	@if ! find data -name "$(GEOIP_DB)" -exec mv {} data/ \; 2>/dev/null; then \
		echo "Error: Database file not found in archive."; \
		echo "Contents of data directory:"; \
		ls -la data/; \
		rm -f data/geoip.tar.gz; \
		exit 1; \
	fi
	@rm -f data/geoip.tar.gz
	@echo "Database downloaded successfully!"
	@echo "Database file size: $$(du -h data/$(GEOIP_DB) | cut -f1)"

# 更新 GeoIP 数据库
update-geoip:
	@echo "Updating GeoIP database..."
	@if [ -z "$(LICENSE_KEY)" ]; then \
		echo "Error: LICENSE_KEY is not set. Please use: make update-geoip LICENSE_KEY=your_key"; \
		exit 1; \
	fi
	@if [ -f "data/$(GEOIP_DB)" ]; then \
		mv data/$(GEOIP_DB) data/$(GEOIP_DB).backup; \
		echo "Created backup: data/$(GEOIP_DB).backup"; \
	fi
	@$(MAKE) download-geoip
	@echo "GeoIP database updated successfully!"

# 下载 ip2region 数据库
download-ip2region:
	@echo "Downloading ip2region database..."
	@mkdir -p data/ip2region
	@if ! go run -v ./scripts/download_ip2region.go; then \
		echo "Error: Failed to download ip2region database. Please check your network connection."; \
		exit 1; \
	fi
	@echo "ip2region database downloaded successfully!"
	@echo "Database file size: $$(du -h data/ip2region/$(IP2REGION_DB) | cut -f1)"

# 更新 ip2region 数据库
update-ip2region:
	@echo "Updating ip2region database..."
	@if [ -f "data/ip2region/$(IP2REGION_DB)" ]; then \
		mv data/ip2region/$(IP2REGION_DB) data/ip2region/$(IP2REGION_DB).backup; \
		echo "Created backup: data/ip2region/$(IP2REGION_DB).backup"; \
	fi
	@$(MAKE) download-ip2region
	@echo "ip2region database updated successfully!"

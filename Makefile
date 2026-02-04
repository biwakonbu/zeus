.PHONY: build clean test install dev dashboard-deps dashboard-dev dashboard-build build-all clean-dashboard storybook storybook-build generate

BINARY_NAME=zeus
VERSION=1.0.0
DASHBOARD_DIR=zeus-dashboard

# go generate を実行（.claude/ からテンプレートファイルをコピー）
generate:
	go generate ./internal/generator/...

# Go ビルド（generate を先に実行）
build: generate
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)

test:
	go test -v ./...

install:
	go install -ldflags "-X main.version=$(VERSION)" .

dev:
	go run . $(ARGS)

# ダッシュボード関連
dashboard-deps:
	cd $(DASHBOARD_DIR) && npm install --legacy-peer-deps

dashboard-dev:
	cd $(DASHBOARD_DIR) && npm run dev

dashboard-build:
	cd $(DASHBOARD_DIR) && npm run build
	mkdir -p internal/dashboard/build
	cp -r $(DASHBOARD_DIR)/build/* internal/dashboard/build/

dashboard-clean:
	rm -rf $(DASHBOARD_DIR)/build $(DASHBOARD_DIR)/.svelte-kit
	rm -rf internal/dashboard/build
	mkdir -p internal/dashboard/build
	echo "placeholder" > internal/dashboard/build/.gitkeep

# Storybook
storybook:
	cd $(DASHBOARD_DIR) && npm run storybook

storybook-build:
	cd $(DASHBOARD_DIR) && npm run build-storybook

# 統合ビルド
build-all: dashboard-build build

# 開発サーバー起動（Go + Vite 並行）
dev-dashboard:
	@echo "Starting Go server in dev mode..."
	@echo "Run 'make dashboard-dev' in another terminal for HMR"
	go run . dashboard --dev --port 8080

# クリーン（全て）
clean-all: clean dashboard-clean

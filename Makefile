PHONY: run bundle_assets

install_deps:
	npm install --dev
	npm install --prod

compile_jsx:
	npm run build

bundle_assets:
	go-bindata -prefix resources -fs -o pkg/assets/assets.go -pkg assets resources/...

run: compile_jsx bundle_assets
	go run . -db.file harvester.db -debug

build: compile_jsx bundle_assets
	go build -o harvester
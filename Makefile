PHONY: run bundle_assets

install_deps:
	npm install --dev
	npm install --prod

compile_jsx:
	npm run build

compile_jsx_dev:
	npm run debug

bundle_assets:
	go-bindata -prefix resources -fs -o pkg/assets/assets.go -pkg assets -ignore DS_Store resources/...

run: compile_jsx_dev bundle_assets
	go run . -db.file harvester.db -debug

build: compile_jsx bundle_assets
	go build -o harvester

package:
	astilectron-bundler -c bundler.config.json -v
	rm bind_*.go
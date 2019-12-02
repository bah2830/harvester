PHONY: run bundle_assets

install_deps:
	npm install @babel/core @babel/cli
	npm install @babel/plugin-transform-react-jsx

compile_jsx:
	./node_modules/.bin/babel --plugins @babel/plugin-transform-react-jsx react -o resources/js/app.js

bundle_assets:
	go-bindata -prefix resources -fs -o pkg/assets/assets.go -pkg assets resources/...

run: compile_jsx bundle_assets
	go run . -db.file harvester.db -debug

build: compile_jsx bundle_assets
	go build -o harvester
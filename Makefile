PHONY: run bundle_assets

run: bundle_assets
	go run . -db.file harvester.db -debug

bundle_assets:
	go-bindata -o assets.go  resources/css/... resources/js/...


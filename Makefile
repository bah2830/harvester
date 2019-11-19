PHONY: build package

NAME=harvester
BUILD_PATH=./artifacts/
ICON_PATH=../icons/icon.png

package_all: package_darwin package_linux package_windows

package_darwin: build
	$(eval OS = darwin)
	$(eval OS_PACKAGE = ${NAME}.app)
	$(call package)

package_linux: build
	$(eval OS = linux)
	$(eval OS_PACKAGE = ${NAME}.tar.gz)
	$(call package)

package_windows: build
	$(eval OS = windows)
	$(eval OS_PACKAGE = ${NAME}.exe)
	$(call package)

package_migrations:
	cd migrations; go-bindata -o ./migrations.go -pkg migrations .; cd ..

build: package_migrations
	go build -o ${BUILD_PATH}${NAME}

generate_icons:
	fyne bundle -package icons -prefix Resource ./icons > ./icons/icons.go

define package
	cd artifacts; fyne package -name ${NAME} -executable ${NAME} -os ${OS} -icon ${ICON_PATH}; cd ..
endef
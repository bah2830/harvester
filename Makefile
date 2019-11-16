PHONY: build package

NAME=harvester
BUILD_PATH=./artifacts/
ICON_PATH=./icon.png

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

build:
	go build -o ${BUILD_PATH}${NAME}

define package
	fyne package -name ${NAME} -executable ${BUILD_PATH}${NAME} -os ${OS} -icon ${ICON_PATH}
	rm -rf ${BUILD_PATH}${NAME}.app
	mv ${OS_PACKAGE} ${BUILD_PATH}
endef
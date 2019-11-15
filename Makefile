PHONY: build package

NAME=harvester
BUILD_PATH=./artifacts/
ICON_PATH=./icon.png

build:
	go build -o ${BUILD_PATH}${NAME}

package: build
	fyne package -name ${NAME} -executable ${BUILD_PATH}${NAME} -os darwin -icon ${ICON_PATH}
	rm -rf ${BUILD_PATH}${NAME}.app
	mv ${NAME}.app ${BUILD_PATH}
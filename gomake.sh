#!/bin/sh
set -e

update() {
    if [ -z ${GOPATH+x} ]; then
        echo -e "\033[0;31mPlease set \$GOPATH\033[0m"
        exit 1
    fi
    url=$(curl -s https://api.github.com/repos/n0rad/gomake/releases | grep browser_download_url | grep 'tar.gz"' | grep `go env GOOS`-`go env GOARCH` | head -n1 | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' | cut -f2 -d' ' | tr -d '"')
    if [ -z ${url} ]; then
        echo "\033[0;31mCannot find a version to download\033[0m"
        exit 1
    fi

	wget -O /tmp/gomake.tar.gz ${url}
	tar xvzf /tmp/gomake.tar.gz -C /tmp
	mkdir -p ${GOPATH}/bin/
	mv /tmp/gomake*/gomake ${GOPATH}/bin/
	rm -Rf /tmp/gomake*

	echo "\033[0;35mVersion downloaded :\033[0m"
	gomake version
}

command -v gomake >/dev/null 2>&1 || {
    echo -e "\033[0;35mGomake not found, downloading\033[0m"
    update
}


gomake $@

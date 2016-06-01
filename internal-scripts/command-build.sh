#!/bin/bash
set -e
set -x
start=`date +%s`
dir=.

#app=gomake
osarchi="$(go env GOHOSTOS)-$(go env GOHOSTARCH)"
[ -z "$1" ] || osarchi="$1"
[ ! -z ${version+x} ] || version="0"

[ -f ${GOPATH}/bin/godep ] || go get github.com/tools/godep
[ -f /usr/bin/upx ] || (echo "upx is required to build" && exit 1)

echo -e "\033[0;32mSave Dependencies\033[0m"
godep save ./${dir}/... || true


# binary
if [ -d ${dir}/internal-scripts ]; then
    [ -f ${GOPATH}/bin/go-bindata ] || go get -u github.com/jteeuwen/go-bindata/...
    go-bindata -nomemcopy -pkg dist -o ${dir}/dist/bindata.go ${dir}/internal-scripts/...
fi

IFS=',' read -ra current <<< "$osarchi"
for e in "${current[@]}"; do
    echo -e "\033[0;32mBuilding $e\033[0m"

    GOOS="${e%-*}" GOARCH="${e#*-}" \
    godep go build -ldflags "-X main.BuildTime=`date -u '+%Y-%m-%d_%H:%M:%S_UTC'` -X main.Version=${version}-`git rev-parse HEAD`" \
        -o $dir/dist/${app}-v${version}-${e}/${app}

    if [ "${e%-*}" != "darwin" ]; then
        echo -e "\033[0;32mCompressing ${e}\033[0m"
        upx ${dir}/dist/${app}-v${version}-${e}/${app} &> /dev/null
    fi

    if [ "${e%-*}" == "windows" ]; then
        mv ${dir}/dist/${app}-v${version}-${e}/${app} ${dir}/dist/${app}-v${version}-${e}/${app}.exe
    fi
done

echo -e "\033[0;32mInstalling\033[0m"

cp ${dir}/dist/${app}-v${version}-$(go env GOHOSTOS)-$(go env GOHOSTARCH)/${app}* ${GOPATH}/bin/

echo -e "\033[0;35mBuild duration : $((`date +%s`-start))s\033[0m"

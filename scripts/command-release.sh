#!/bin/bash
set -x
dir=$( dirname "$0" )/..

if [ $# != 2 ]; then
    echo "Usage: release.sh VERSION GITHUB_KEY"
    exit 1
fi

app="gomake"
github_repo="n0rad/gomake"
version=$1
access_token=$2
osarchi="linux-amd64,darwin-amd64,windows-amd64"

require_clean_work_tree () {
    # Update the index
    git update-index -q --ignore-submodules --refresh
    err=0

    # Disallow unstaged changes in the working tree
    if ! git diff-files --quiet --ignore-submodules --
    then
        echo >&2 "cannot $1: you have unstaged changes."
        git diff-files --name-status -r --ignore-submodules -- >&2
        err=1
    fi

    # Disallow uncommitted changes in the index
    if ! git diff-index --cached --quiet HEAD --ignore-submodules --
    then
        echo >&2 "cannot $1: your index contains uncommitted changes."
        git diff-index --cached --name-status -r --ignore-submodules HEAD -- >&2
        err=1
    fi

    if [ $err = 1 ]
    then
        echo >&2 "Please commit or stash them."
        exit 1
    fi
}

${dir}/clean.sh
version=${version} ${dir}/build.sh ${osarchi}
${dir}/test.sh
require_clean_work_tree

echo -e "\033[0;32mCompress releases\033[0m"
cd ${dir}/dist
for i in *-* ; do
    if [ -d "$i" ]; then
        tar czf ${i}.tar.gz ${i}
    fi
done
cd -

git tag v${version} -a -m "Version $version"
git push --tags
#
#sleep 5
#
posturl=$(curl --data "{\"tag_name\": \"v${version}\",\"target_commitish\": \"master\",\"name\": \"v${version}\",\"body\": \"Release of version ${version}\",\"draft\": false,\"prerelease\": false}" https://api.github.com/repos/${github_repo}/releases?access_token=${access_token} | grep "\"upload_url\"" | sed -ne 's/.*\(http[^"]*\).*/\1/p')

for i in ${dir}/dist/*.tar.gz ; do
    fullpath=$(ls ${i})
    filename=${fullpath##*/}
    curl -i -X POST -H "Content-Type: application/x-gzip" --data-binary "@${fullpath}" "${posturl%\{?name,label\}}?name=${filename}&label=${filename}&access_token=${access_token}"
done

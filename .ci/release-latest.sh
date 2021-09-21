#!/bin/bash

ROOT_DIR=$(dirname $(dirname $(realpath -m ${0})))

github(){
  hub -C $ROOT_DIR ${@}
}

rel(){
  ver="$(git tag -l --points-at HEAD | grep -v latest | head -n1)"
  if [[ "${ver}" == "" ]]; then
    echo latest
  else
    echo "${ver}"
  fi
}

latest="$(rel)"

refresh(){
  git push --delete origin "$latest" || true
  git tag -d "latest" || true
  git tag -d "$latest" || true
  git tag "$latest" HEAD
  git push origin "$latest"
  github release delete "$latest"
}

upload(){
  github release | grep "$latest" || github release create -m "$latest" "$latest"
  github release edit -m "$latest" "$latest" -a "$1"
}

loop_unix(){
  cd ${ROOT_DIR}/bin
  ls -1d {android,darwin,linux,*bsd}/* | while read dir; do
    pushd $dir
    compressed="${OLDPWD}/k0s-${dir////-}.tar.gz"
    tar cz * > "$compressed"
  # upload $(realpath -m $compressed)
    popd
  done
}

loop_windows(){
  cd ${ROOT_DIR}/bin
  ls -1d windows/* | while read dir; do
    pushd $dir
    compressed="${OLDPWD}/k0s-${dir////-}.zip"
    zip - * > "$compressed"
  # upload $(realpath -m $compressed)
    popd
  done
}

main(){
  loop_unix
  loop_windows
}

main

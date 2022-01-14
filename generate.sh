#!/usr/bin/bash
archs=(amd64 arm64)

for arch in ${archs[@]}
do
        env GOOS=linux GOARCH=${arch} go build -o validator_${arch}
done

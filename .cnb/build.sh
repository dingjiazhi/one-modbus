#!/bin/bash
set -ex

VER="${CNB_BRANCH#v}"
FNAME="modbusrtu_broker_${VER}.exe"

P1=$(echo $VER | cut -d. -f1)
P2=$(echo $VER | cut -d. -f2)
P3=$(echo $VER | cut -d. -f3)

printf '{"FixedFileInfo":{"FileVersion":{"Major":%s,"Minor":%s,"Patch":%s,"Build":0},"ProductVersion":{"Major":%s,"Minor":%s,"Patch":%s,"Build":0},"FileFlagsMask":"3f","FileFlags":"00","FileOS":"040004","FileType":"01","FileSubType":"00"},"StringFileInfo":{"Comments":"get: gitee.com/dingjiazhi/one-modbus/releases","CompanyName":"DJ","FileDescription":"modbusrtu_broker","FileVersion":"%s.0","InternalName":"modbusrtu_broker","LegalCopyright":"Copyright 2026 DJ","OriginalFilename":"modbusrtu_broker.exe","ProductName":"one-modbus","ProductVersion":"%s.0"},"IconPath":"app.ico","ManifestPath":""}' \
  $P1 $P2 $P3 $P1 $P2 $P3 $VER $VER > /tmp/versioninfo.json
~/go/bin/goversioninfo -icon app.ico /tmp/versioninfo.json

export GOPROXY=https://goproxy.cn,direct
go mod init modbusrtu_broker 2>/dev/null
go mod tidy
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
  CGO_CFLAGS="-D_WIN32" CGO_LDFLAGS="-lws2_32" \
  go build -ldflags '-s -w -extldflags=-static' -o /tmp/$FNAME

ls -lh /tmp/$FNAME
echo "BUILD SUCCESS on CNB: $FNAME"

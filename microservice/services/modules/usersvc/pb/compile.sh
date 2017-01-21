#!/usr/bin/env sh

# Install proto3 from source
#  brew install autoconf automake libtool
#  git clone https://github.com/google/protobuf
#  ./autogen.sh ; ./configure ; make ; make install
#
# Update protoc Go bindings via
#  go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
#
# See also
#  https://github.com/grpc/grpc-go/tree/master/examples
#  https://github.com/grpc/grpc/blob/master/INSTALL.md   protoc-gen-php install

protoc svc.proto --go_out=plugins=grpc:.
protoc --php_out=:. --grpc_out=:. --plugin=protoc-gen-grpc=../../../vendor/grpc/bins/opt/grpc_php_plugin svc.proto

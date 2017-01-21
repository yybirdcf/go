<?php

/*

$ git clone https://github.com/grpc/grpc
Build and install the gRPC C core library

$ cd grpc
$ git pull --recurse-submodules && git submodule update --init --recursive
$ make
$ sudo make install
gRPC PHP extension

Compile the gRPC PHP extension

$ cd grpc/src/php/ext/grpc
$ phpize
$ ./configure
$ make
$ sudo make install

*/

error_reporting(E_ALL & ~E_NOTICE & ~E_WARNING);
ini_set('display_errors', 'On');

require_once __DIR__ . '/../../../../vendor/grpc/src/php/vendor/autoload.php';
require_once __DIR__ . '/../../pb/GPBMetadata/Svc.php';
require_once __DIR__ . '/../../pb/svc_grpc_pb.php';
require_once __DIR__ . '/../../pb/Pb/GetUserinfoRequest.php';
require_once __DIR__ . '/../../pb/Pb/GetUserinfoResponse.php';
require_once __DIR__ . '/../../pb/Pb/Userinfo.php';

use Google\Protobuf\Internal\RepeatedField;
use Google\Protobuf\Internal\GPBType;

if($argc < 2)
{
  echo "php main.php [rpc host: 127.0.0.1:8082]";
  return;
}

$host = $argv[1];

$client = new \Pb\UsersvcClient($host, [
  'credentials' => Grpc\ChannelCredentials::createInsecure()
  ]);

$request = new \Pb\GetUserinfoRequest();
$request->setId(97);
list($response, $status) = $client->GetUserinfo($request)->wait();
if($status->code == 0)
{
  var_dump($response->getUserinfo()->getId());
}

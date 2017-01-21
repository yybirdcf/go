<?php
// GENERATED CODE -- DO NOT EDIT!

namespace Pb {

  // user information service definition
  //
  class UsersvcClient extends \Grpc\BaseStub {

    /**
     * @param string $hostname hostname
     * @param array $opts channel options
     * @param Grpc\Channel $channel (optional) re-use channel object
     */
    public function __construct($hostname, $opts, $channel = null) {
      parent::__construct($hostname, $opts, $channel);
    }

    /**
     * get user basic information
     * @param \Pb\GetUserinfoRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     */
    public function GetUserinfo(\Pb\GetUserinfoRequest $argument,
      $metadata = [], $options = []) {
      return $this->_simpleRequest('/pb.Usersvc/GetUserinfo',
      $argument,
      ['\Pb\GetUserinfoResponse', 'decode'],
      $metadata, $options);
    }

  }

}

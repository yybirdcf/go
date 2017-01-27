CREATE TABLE `message` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `ver` int(11) NOT NULL DEFAULT '0' COMMENT '消息版本号',
  `mt` int(11) NOT NULL DEFAULT '0' COMMENT '消息类型',
  `mid` bigint(20) NOT NULL DEFAULT '0' COMMENT '消息id',
  `sid` bigint(20) NOT NULL DEFAULT '0' COMMENT '发送者',
  `rid` bigint(20) NOT NULL DEFAULT '0' COMMENT '接收者',
  `ext` text NOT NULL COMMENT '扩展属性',
  `pl` text NOT NULL COMMENT 'payload内容',
  `ct` bigint(20) NOT NULL DEFAULT '0' COMMENT '创建时间，ms',
  PRIMARY KEY (`id`),
  KEY `idx_rid_mid` (`rid`,`mid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

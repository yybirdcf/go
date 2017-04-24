CREATE TABLE `service` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `service` varchar(255) NOT NULL DEFAULT '' COMMENT '服务名',
  `iport` varchar(255) NOT NULL DEFAULT '' COMMENT 'ip:port',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `ping_uri` varchar(255) NOT NULL DEFAULT '' COMMENT 'ping uri, 如: /index',
  `ping_host` varchar(255) NOT NULL DEFAULT '' COMMENT 'ping host, 如: api.demo.com',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_service_iport` (`service`,`iport`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `route` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `rule` varchar(255) NOT NULL DEFAULT '' COMMENT '规则',
  `service` varchar(255) NOT NULL DEFAULT '' COMMENT '服务名',
  `cache` int(11) NOT NULL DEFAULT '0' COMMENT '默认不缓存，缓存时间s',
  `sign` int(11) NOT NULL DEFAULT '0' COMMENT '默认不验证sign',
  `auth` int(11) NOT NULL DEFAULT '0' COMMENT '默认不需要登录态',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '描述',
  `create_time` int(11) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `update_time` int(11) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '标签名字，唯一标示规则',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_rule` (`rule`)
) ENGINE=InnoDB AUTO_INCREMENT=45 DEFAULT CHARSET=utf8;

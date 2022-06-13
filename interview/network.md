## HTTPS

### 加密过程

1. 自动连接到server的443端口
2. server端生成公钥和私钥
3. server将数字证书（公钥）发送给client，自己留下私钥
4. client收到server的数字证书之后，验证该数字证书是否有效，比如颁发机构，过期时间等，如果有效，就会生成一个随机值
5. client将随机值放入公钥里面，传输给server
6. server使用私钥进行解密，获取里面的随机值
7. server将随机值和私钥进行对称加密
8. server将对称加密的内容发送给client
9. client使用自己私钥进行解密，获取内容
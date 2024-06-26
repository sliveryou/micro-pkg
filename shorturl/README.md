# 短地址标识符生成 shorturl

短地址标识符生成。
   - authored by sliveryou

## 原理

当我们在浏览器里输入 `https://xxx.xxx/RlB2PdD` 时，DNS 首先解析获得 `https://xxx.xxx` 的 IP 地址。  
当 DNS 获得 IP 地址以后（比如：`192.168.10.1`），会向这个地址发送 HTTP GET 请求，查询短地址标识符 `RlB2PdD` 在服务器中记录的对应的长 URL，然后请求通过 HTTP 301 的 Location 响应头转到对应的长 URL。    
目前比较流行的算法有两种：自增序列算法和摘要算法。

## 自增序列算法

设置 ID 自增，一个 10 进制 ID 对应一个 62 进制的数值，1 对 1，也就不会出现重复的情况。  
这个利用的就是低进制转化为高进制时，字符数会减少的特性。  
短地址标识符的长度一般设为 6 位，而每一位是由 `[0-9,A-Z,a-z]` 总共 62 个字母组成的，所以 6 位的话，总共会有 $62^6$ ~= 568 亿种组合，基本上够用了。

## 摘要算法

将长 URL 用 MD5 生成 32 位签名串，分为 4 段，每段 8 个字节。  
对这四段循环处理, 取 8 个字节，将它看成 16 进制串与 0x3fffffff（30位1）与操作, 即超过 30 位的忽略处理。  
这 30 位分成 6 段，每 5 位的数字作为字母表的索引取得特定字符，依次进行获得 6 位字符串。  
总的 MD5 串可以获得 4 个 6 位串，取里面的任意一个就可作为这个长 URL 的短地址标识符。

## 对比

第一种算法的好处就是简单好理解，永不重复。  
但是短地址标识符的长度不固定，随着 ID 变大从一位长度开始递增。如果非要让短地址标识符长度固定也可以让 ID 从指定数字开始递增。  
另外，该算法是和 ID 绑定的，如果允许自定义短地址标识符就会占用之后的短码，之后的 ID 要生成短码的时候就会发现短码已经被用了，那么 ID 自增一对一不冲突的优势就体现不出来了。

第二种算法，存在碰撞（重复）的可能性，虽然几率很小，但生成的短地址标识符位数是比较固定的，不会从一位长度递增到多位。

## 总结

本模块采用方式二的方法，将长 URL 基于 Murmur3 生成校验和，对校验和进行 n 次移位，从而获得 n 位的短地址标识符。

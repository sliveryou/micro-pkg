# gstream

通用 gRPC 流式消息内容读写器。
   - authored by sliveryou

## 背景

在一份 gRPC message 消息的定义中，往往会有较大体量的元数据：

```protobuf
// UploadFileReq 上传文件请求
message UploadFileReq {
  int64 user_id = 1;
  string file_name = 2;
  string file_type = 3;
  string file_hash = 4;
  int64 file_size = 5;
  bytes file_data = 6;
}

// 其中 file_data 就是较大的文件二进制数据
```

在 `google.golang.org/grpc@v1.29.1/server.go` 中，服务端接收的最大消息字节数的被设置为了 4MB，且其他语言的 gRPC 消息接收限制大抵也是如此：

```go
const (
	defaultServerMaxReceiveMessageSize = 1024 * 1024 * 4
)
```

如果想传输较大体量的消息，一般有两种策略：

1. 修改消息的阈值：  
   通过 `grpc.MaxCallRecvMsgSize(bytes int)` 和 `grpc.MaxCallSendMsgSize(bytes int)` 设置，最大不能超过 2GB
2. 流式消息传输：  
   在对应 rpc 定义，将较大体量的消息前添加 `stream` 关键字，如 `rpc UploadFile (stream UploadFileReq) returns (UploadFileResp); // UploadFile 上传文件`

gRPC 官方建议一次消息传输的最大字节不超过 4MB，所以传输较大消息时，最好是选择流式传输。

## 设计思路

一般的客户端消息流式发送：

```go
// 伪代码

func cliDemo() {
	f, err := os.Open("test.pdf")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	cli, err := flieclient.UploadFile(context.Background())
	if err != nil {
		panic(err)
	}
	defer cli.CloseSend()

	// 首次发送，不传输文件内容，只传输文件相关信息
	err = cli.Send(&fileclient.UploadFileReq{
		FileName: f.Filename,
		FileType: f.FileType,
		FileHash: f.FileHash,
		FileSize: f.Size,
	})
	if err != nil {
		panic(err)
	}

	// 定义一个 buf，从源文件不断读取数据到 buf，而后发送消息
	chunkSize := (3 << 20) + (1 << 19) // 3.5MB
	buf := make([]byte, chunkSize)
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		// 后续发送，不传输文件相关信息，只传输文件内容
		err = cli.Send(&fileclient.UploadFileReq{FileData: buf[:n]})
		if err != nil {
			panic(err)
		}
	}

	resp, err := cli.CloseAndRecv()
	if err != nil {
		panic(err)
	}
}
```

一般的服务端消息流式接收：

```go
// 伪代码

func svrDemo(svr file.File_UploadFileServer) {
	// 首次接收，获取文件相关信息
	fi, err := svr.Recv()
	if err != nil {
		panic(err)
	}

	// 定义一个 buf，不断将消息中的元数据写入其中
	var buf bytes.Buffer
	for {
		req, err := svr.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		_, err := io.Copy(buf, bytes.NewBuffer(req.FileData))
		if err != nil {
			panic(err)
		}
	}
	
	// 后续操作
	key := fmt.Sprintf("common/%s.%s", fi.FileHash, fi.FileType)
	oss.PutObject(key, buf)
}
```

主要构建逻辑：
1. 客户端不断从源数据读取一定的数据分块，再构建成消息体，Send 消息到服务端，直到数据读完了，则发送 CloseAndRecv 信号，等待服务端的回复；
2. 服务端不断 Recv 消息，抽取出其中的分块元数据，可以将其整合到 buf 中或者临时文件里，等待下一步处理。

这种模式的主要缺点：

1. 会发现流式传输好像都是一样代码逻辑，但是却具有业务的特征（特定消息结构体，业务相关），无法单独抽象出来
2. 服务端每一次在循环中接收都是完整的消息结构，然后抽取其中的元数据将其转化成 io.Reader，给相关 io.Writer 调用，  
要知道，一般需要流式传输的数据往往是较大的文件二进制数据，如很大的视频或者图片等，为了在 gRPC 中传输，被客户端切割，  
然后被服务端接收所拼凑还原，在拼凑还原的过程中，存在一个中间态，是把前部分的数据放在内存里呢？
还是生成一个临时文件，将数据存放在其中呢？
   
在对接文件对象存储的业务中，我设计了一个 OSS 服务通用接口来对接阿里云、华为云和腾讯云的对象存储服务：

```go
// OSS OSS 服务接口
type OSS interface {
	// Cloud 获取云服务商名称
	Cloud() string
	// GetUrl 获取对象在 OSS 上的完整访问 URL
	GetUrl(key string) string
	// PutObject 上传对象至 OSS
	PutObject(key string, reader io.Reader) (string, error)
	// DeleteObjects 批量删除 OSS 上的对象
	DeleteObjects(keys ...string) error
	// UploadFile 上传文件至 OSS，filePath：文件路径，partSize：分块大小（字节），routines：并发数
	UploadFile(key, filePath string, partSize int64, routines int) (string, error)
	// AuthorizedUpload 授权上传至 OSS，expires：过期时间（秒）
	AuthorizedUpload(key string, expires int) (string, error)
}
```

将上传对象至 OSS 设计成 `PutObject(key string, reader io.Reader) (string, error)`   
而不是 `PutObject(key string, buf []bytes) (string, error)` 的原因显而易见：边收边传尽量减少中间态才是更好的传输方案。

## 使用说明

客户端消息流式发送：

```go
// 伪代码

func cliDemo() {
	f, err := os.Open("test.pdf")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	
	cli, err := flieclient.UploadFile(context.Background())
	if err != nil {
		panic(err)
	}
	defer cli.CloseSend()
	
	 // 首次发送，不传输文件内容，只传输文件相关信息
	err = cli.Send(&fileclient.UploadFileReq{
		FileName: f.Filename,
		FileType: f.FileType,
		FileHash: f.FileHash,
		FileSize: f.Size,
	})
	if err != nil {
		panic(err)
	}
	
	chunkSize := (3 << 20) + (1 << 19) // 3.5MB
	// 新建gRPC流式消息内容写入器，传入客户端对象、消息请求体对象、指定传输消息字段和传输消息块大小
	writer := gstream.MustNewStreamWriter(cli, &fileclient.UploadFileReq{}, "FileData", chunkSize)
	_, err = io.Copy(writer, f)
	if err != nil {
		panic(err)
	}
	err = writer.Close()
	if err != nil {
		panic(err)
	}
	
	resp, err := cli.CloseAndRecv()
	if err != nil {
		panic(err)
	}
}
```

服务端消息流式接收：

```go
// 伪代码

func svrDemo(svr file.File_UploadFileServer) {
	// 首次接收，获取文件相关信息
	fi, err := svr.Recv()
	if err != nil {
		panic(err)
	}
	
	key := fmt.Sprintf("common/%s.%s", fi.FileHash, fi.FileType)
	// 新建gRPC流式消息内容读取器，传入服务端对象、消息请求体对象、指定接收消息字段和总计消息块大小
	reader := gstream.MustNewStreamReader(svr, &file.UploadFileReq{}, "FileData", fi.FileSize)
	oss.PutObject(key, reader)
}
```

## 实现原理

- StreamWriter
   - 内部传入 gRPC 客户端流对象，利用反射动态创建消息对象，并对指定 []byte 字段赋值
   - 内部申请 chunkSize 大小的缓存区，当缓存区写满时再调用客户端流对象进行消息 Send
   - Close 时将不足 chunkSize 大小缓存区数据全部写入消息体，进行最后一次发送

- StreamReader
   - 内部传入 gRPC 服务端流对象，利用反射动态创建消息对象，并对指定 []byte 字段取值
   - 内部将每次读取的消息体进行缓存，直到外界将本次消息体的内容读完时，再进行消息 Recv
   - 消息全部读完时，返回 io.EOF，让外部调用知晓数据已读取完毕

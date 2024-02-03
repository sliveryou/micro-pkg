# 接口错误码

> 注意：HTTP 状态码为 `200`，错误码为 `0` 时，才代表接口请求成功

| **错误** | **错误码** | **释义** | **HTTP 状态码** |
|:---------|:-----------|:---------|:----------------|
| OK | 0 | ok | <font color='green'>200</font> |
| ErrCommon | 97 | 通用错误 | <font color='green'>200</font> |
| ErrRecordNotFound | 98 | 记录不存在 | <font color='green'>200</font> |
| ErrUnexpected | 99 | 服务器繁忙，请稍后重试 | <font color='green'>200</font> |
| ErrInvalidParams | 100 | 请求参数错误 | <font color='green'>200</font> |

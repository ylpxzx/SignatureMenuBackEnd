# SignatureMenuBackEnd

## 数据补丁

### 生成模拟菜谱数据

在 `SignatureMenuBackEnd` 目录下执行：

```bash
go run main.go patch run_mock_data
```

该命令会创建或复用模拟账号 `mock_data / mock123456`，清理该账号下已有的模拟菜谱，并随机生成 50 条真实感菜谱数据。数据会写入当前后端配置使用的 `DATA_FILE`。

# kitex_gorm

## Introduce

A demo with `Kitex` and `Gorm`

- Use `thrift` IDL to define `RPC` interface
- Use `kitex` to generate code
- Use `Gorm/Gen` and `MySQL`


- `/cmd` generate gorm_gen
- `/dao` initialize mysql connection
- `/model` gorm generated user struct
- `/hander.go` basic biz logic for updateUser, addUser, deleteUser, queryUser


## IDL

This demo use `thrift` IDL to define `RPC` interface. The specific interface define in [user.thrift](idl/user.thrift)

## Code generation tool

This demo use `kitex` to generate code. The use of `kitex` refers
to [kitex](https://www.cloudwego.io/docs/kitex/tutorials/code-gen/)

The `kitex` commands used can be found in [Makefile](Makefile)

## GORM/Gen

GEN: Friendly & Safer GORM powered by Code Generation.

This demo use `GORM/Gen` to operate `MySQL` and refers to [Gen](https://gorm.io/gen/index.html).

#### Quick Start

- Update the Database DSN to your own in [Database init file](internal/repository/mysql/init.go).
- Refer to the code comments, write the configuration in [Generate file](cmd/generate.go).
- Using the following command for code generation, you can generate structs from databases or basic type-safe DAO API for struct.
```bash
cd bizdemo/kitex_gorm_gen/cmd
go run generate.go
```
- For more Gen usage, please refer to [Gen Guides](https://gorm.io/gen/index.html).

## How to run

### Run mysql docker

```bash
cd bizdemo/kitex_gorm_gen && docker-compose up
```

### Run demo

```bash
cd bizdemo/kitex_gorm_gen
go build -o kitex_gorm_gen && ./kitex_gorm_gen
```

# structure

以下是一个典型的后端 Golang 项目目录结构，它考虑了你列出的技术栈以及未来可能添加的其他组件：

```
project_root/
├── api/                # API 定义文件 (例如, .proto 文件)
│   └── v1/
│       └── service.proto
├── cmd/                # 项目的主入口点
│   ├── server/         # Kitex 服务器主程序
│   │   └── main.go
│   └── client/         # 示例 Kitex 客户端主程序
│       └── main.go
├── config/             # 配置文件
│   ├── config.go       # 配置文件读取逻辑
│   └── config.yaml     # 默认配置文件
├── internal/           # 项目内部私有代码
│   ├── app/            # 应用层，包含业务逻辑
│   │   ├── module1/    # 模块 1
│   │   │   ├── handler.go # Kitex handler
│   │   │   ├── service.go # 业务逻辑
│   │   │   └── repository.go # 数据访问接口
│   │   └── module2/    # 模块 2
│   │       └── ...
│   ├── cache/          # 缓存层
│   │   ├── local/      # 本地缓存实现
│   │   │   └── local.go
│   │   └── redis/      # Redis 缓存实现
│   │       └── redis.go
│   ├── client/         # 客户端实现
│   │   ├── elasticsearch/ # Elasticsearch 客户端
│   │   │   └── es.go
│   │   ├── gpt/        # GPT 客户端
│   │   │   └── gpt.go
│   │   └── kitex/      # Kitex 客户端
│   │       └── kitex_client.go
│   ├── mq/             # 消息队列
│   │   └── consumer.go # 消息消费者
│   ├── pkg/            # 内部可复用包
│   │   └── utils/      # 工具函数
│   │       └── ...
│   └── repository/     # 数据访问层实现
│       └── mysql/      # MySQL 数据访问实现
│           └── gen.go  # Gorm/Gen 生成的代码
├── pkg/                # 对外公开的包
│   └── ...
├── scripts/            # 脚本文件 (例如, 编译脚本, 部署脚本)
│   ├── build.sh
│   └── deploy.sh
├── test/               # 测试文件
│   ├── integration/    # 集成测试
│   └── unit/           # 单元测试
├── .gitignore          # Git 忽略文件
├── go.mod              # Go 模块依赖
├── go.sum              # Go 模块校验和
└── README.md           # 项目说明文档
```

**说明:**

*   **api/**: 存放 API 定义文件，例如 `.proto` 文件。如果使用其他协议，例如 HTTP，可以根据需要调整。
*   **cmd/**: 存放项目的主入口点，通常每个可执行文件一个目录。
*   **config/**: 存放配置文件和配置读取逻辑。
*   **internal/**: 存放项目内部私有代码，这些代码不应该被外部项目导入。
    *   **app/**: 应用层，包含业务逻辑。按模块划分，每个模块包含 `handler` (Kitex handler), `service` (业务逻辑), `repository` (数据访问接口)。
    *   **cache/**: 缓存层，包含本地缓存和 Redis 缓存的实现。
    *   **client/**: 存放各种客户端的实现，例如 Elasticsearch, GPT, Kitex 客户端。
    *   **mq/**: 消息队列相关代码，例如消息消费者。
    *   **pkg/**: 内部可复用包，例如工具函数。
    *   **repository/**: 数据访问层实现，例如 MySQL 数据访问实现，包含 Gorm/Gen 生成的代码。
*   **pkg/**: 存放对外公开的包，这些包可以被外部项目导入。
*   **scripts/**: 存放脚本文件，例如编译脚本、部署脚本。
*   **test/**: 存放测试文件，包括集成测试和单元测试。
*   **.gitignore**: Git 忽略文件。
*   **go.mod**: Go 模块依赖文件。
*   **go.sum**: Go 模块校验和文件。
*   **README.md**: 项目说明文档。

**设计原则:**

*   **清晰分离**: 按照功能和职责划分目录，使代码结构清晰，易于维护。
*   **模块化**: 将业务逻辑按模块划分，便于代码复用和管理。
*   **可扩展**: 目录结构易于扩展，可以方便地添加新的组件和功能。
*   **约定优于配置**: 遵循 Golang 社区的通用约定，例如 `internal` 和 `pkg` 目录的使用。

**后续添加组件:**

*   如果需要添加新的组件，例如新的数据库、新的客户端等，可以在 `internal/` 目录下创建相应的目录，例如 `internal/client/mongodb/` 或 `internal/repository/postgres/`。
*   如果需要添加新的业务模块，可以在 `internal/app/` 目录下创建新的模块目录。

这个目录结构是一个通用的参考，你可以根据项目的实际情况进行调整。希望这个结构能帮助你更好地组织你的 Golang 项目!

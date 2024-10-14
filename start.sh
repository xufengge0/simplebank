
#!/bin/bash
# 使用

set -e
# 该命令确保脚本在遇到任何错误时立即退出。

echo "run db migration"

# 执行数据库迁移命令
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up
# 使用 `/app/migrate` 可执行文件运行数据库迁移任务。
# `-path /app/migration`: 指定存储迁移文件的路径为 `/app/migration`。
# `-database "$DB_SOURCE"`: 使用环境变量 `$DB_SOURCE` 作为数据库连接字符串，连接到目标数据库。
# `-verbose`: 以详细模式运行，输出更多的迁移过程日志。
# `up`: 运行数据库迁移的“升级”操作，执行未应用的迁移文件，更新数据库。

echo "start app"

exec "$@"
# 使用 `exec` 来替换当前 shell 进程并执行传递给脚本的命令。



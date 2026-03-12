# Migrations

当前项目的数据库表结构仍由应用启动时自动初始化。

这个目录预留给后续正式迁移脚手架，例如：

- 001_init_tasks.sql
- 002_add_indexes.sql
- 003_add_operator_approver.sql

后续可以接入 migrate / goose 等工具，将数据库初始化逻辑从代码里逐步迁出。

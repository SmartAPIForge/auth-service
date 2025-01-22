# auth-service

This service provides sso functionality in SmartAPIForge backend system

### Get prepared:

##### Fast init

1) Fill .env file in project root (check .env.xmpl as reference)
2) Change dsn args in Taskfile.yaml if needed
2) Up dependencies + run application via ```task init```

##### or

1) Fill .env file in project root (check .env.xmpl as reference)
2) Change dsn args in Taskfile.yaml if needed
3) Raise postgres database via ```task db_raise```
4) Run migrations ```task db_migrate```
5) Seed database ```task db_seed```
6) Run auth-service application ```task run``` (automatically calls ```task build``` before as dependency)
approval-service/
├── cmd/
│ └── main.go
├── config/
│ └── config.go
├── constants/
│ └── constants.go
├── controller/v1/
│ └── slack_controller.go
├── db/ (empty - reserved)
├── document/
│ ├── add_members.txt
│ ├── create_channel.txt
│ ├── folder_structure.md
│ ├── mention_req_res.txt
│ ├── mentions.txt
│ └── req_res.txt
├── middleware/
│ ├── logging.go
│ └── recovery.go
├── models/ (empty - reserved)
├── resources/
│ ├── common.go
│ ├── slack_channel.go
│ ├── slack_channel_info.go
│ ├── slack_member.go
│ ├── slack_message.go
│ └── slack_user.go
├── routes/
│ └── routes.go
├── service/v1/
│ ├── interfaces.go
│ └── slack_service.go
├── utils/ (empty - reserved)
├── validator/
│ ├── create_channel_validator.go
│ └── message_validator.go
├── .env
├── .env.example
├── Dockerfile
├── go.mod
└── go.sum

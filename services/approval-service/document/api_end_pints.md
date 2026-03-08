API Endpoints

Base URL

`http://localhost:8083

Endpoints List

1 GET `/health` Health check
2 POST `/api/v1/slack/channel/create` Create Slack channel
3 GET `/api/v1/slack/channels/all` Get all Slack channels
4 POST `/api/v1/slack/channel/by-name` Get channel by name
5 POST `/api/v1/slack/channel/by-id` Get channel by ID
6 GET `/api/v1/slack/users/all` Get all Slack users
7 POST `/api/v1/slack/user/by-name` Get user by name
8 POST `/api/v1/slack/user/by-id` Get user by ID
9 GET `/api/v1/slack/apps/all` Get all Slack apps
10 POST `/api/v1/slack/member/add` Add member to channel
11 POST `/api/v1/slack/message/send` Send message to channel
12 POST `/api/v1/slack/dm-channel/get` Get DM channel ID
13 POST `/api/v1/slack/approval-form-button/send` Send approval form button to channel
14 GET `/api/v1/approval/all` Get all approval requests
15 GET `/api/v1/approval/pending` Get pending approval requests
16 POST `/api/v1/approval/by-id` Get approval request by ID
17 POST `/api/v1/approval/request` Create approval request
18 POST `/api/v1/approval/domain-change` Create domain change approval request

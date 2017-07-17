# Kindred

> Kindred is a video chat platform that pairs people of varied backgrounds to bridge differences. Each user has a profile with key demographic information used in a proprietary algorithm that pairs different people.  Users are allowed a limited number of pairs per day to ensure meaningful connections.  Discussions are guided by a question of the day to facilitate conversation.  Conversations can be continued via an embedded messaging system if both parties agree.  

## Team

  - __Product Owner__: Zane
  - __Scrum Master__: Nowreen
  - __Development Team Members__: Lindsay, Jon

## Table of Contents

1. [Requirements](#requirements)
1. [Development](#development)
    1. [Installing Dependencies](#installing-dependencies)
    1. [Tasks](#tasks)
1. [Team](#team)
1. [Contributing](#contributing)

## Requirements

- Node
- NPM
- Redis
- PostgreSQL
- Go
- Glide

## Run Locally

1) Install dependencies: From within the root directory:

```sh
npm install
glide install
```

2) In first shell: ```redis-server```
3) In second shell: ```npm run build:watch```
4) In third shell: ```npm run start```
5) In fourth shell: ```node twilio-server/index.js```
6) Open browser to http://localhost:8080

### 
### Roadmap

View the project roadmap [here](https://github.com/KindredApp/kindred/issues?utf8=%E2%9C%93&q=)


## Contributing

See our [contribution guide](https://github.com/KindredApp/kindred/blob/master/CONTRIBUTING.md) for contribution guidelines.

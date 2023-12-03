
# Discord Bot

Backend for [Discord Bot frontend](https://github.com/fdvky1/discord-bot-fe)



## Setup
Make sure you have finished setting up the [Frontend](https://github.com/fdvky1/discord-bot-fe)\
Clone this project

fill the .env

```bash
SUPABASE_URL="your project url"
SUPABASE_KEY="service role key(for bypass RLS)"
POSTGRESQL_URL="postgresql url(if you want to use postgre from supabase please enable the RLS after table created)"
```


    
## Installing & build

```bash
  go mod tidy
  go build -o bot main.go
```

## start
```bash
  ./bot
```
